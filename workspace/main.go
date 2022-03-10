package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/mazrean/genorm"
	orm "github.com/mazrean/genorm-workspace/workspace/genorm"
	"github.com/mazrean/genorm-workspace/workspace/genorm/message"
	"github.com/mazrean/genorm-workspace/workspace/genorm/user"
)

func main() {
	db, err := openDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = migration(db)
	if err != nil {
		panic(err)
	}

	err = runQuery(db)
	if err != nil {
		panic(err)
	}
}

func openDatabase() (*sql.DB, error) {
	user, ok := os.LookupEnv("DB_USERNAME")
	if !ok {
		return nil, errors.New("DB_USERNAME is not set")
	}

	pass, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		return nil, errors.New("DB_PASSWORD is not set")
	}

	host, ok := os.LookupEnv("DB_HOSTNAME")
	if !ok {
		return nil, errors.New("DB_HOSTNAME is not set")
	}

	port, ok := os.LookupEnv("DB_PORT")
	if !ok {
		return nil, errors.New("DB_PORT is not set")
	}

	database, ok := os.LookupEnv("DB_DATABASE")
	if !ok {
		return nil, errors.New("DB_DATABASE is not set")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FTokyo&charset=utf8mb4",
		user,
		pass,
		host,
		port,
		database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

func migration(db *sql.DB) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migration",
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	return nil
}

func runQuery(db *sql.DB) error {
	userID := uuid.New()
	affectedRows, err := genorm.
		Insert(orm.User()).
		Values(&orm.UserTable{
			ID:       userID,
			Name:     genorm.Wrap("user"),
			Password: genorm.Wrap("password"),
		}).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	fmt.Println(affectedRows, userID)

	messageID1 := uuid.New()
	messageID2 := uuid.New()
	_, err = genorm.
		Insert(orm.Message()).
		Values(&orm.MessageTable{
			ID:        messageID1,
			UserID:    userID,
			Content:   genorm.Wrap("hello"),
			CreatedAt: genorm.Wrap(time.Now()),
		}, &orm.MessageTable{
			ID:        messageID2,
			UserID:    userID,
			Content:   genorm.Wrap("world"),
			CreatedAt: genorm.Wrap(time.Now()),
		}).
		Fields(message.ID, message.UserID, message.Content, message.CreatedAt).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}

	userValues, err := genorm.
		Select(orm.User()).
		Fields(user.Name, user.Password).
		Where(genorm.EqLit(user.IDExpr, userID)).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	fmt.Println(userValues)

	userValue, err := genorm.
		Select(orm.User()).
		Fields(user.Name, user.Password).
		Where(genorm.EqLit(user.IDExpr, userID)).
		Get(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	fmt.Println(userValue)

	userNames, err := genorm.
		Pluck(orm.User(), user.NameExpr).
		Distinct().
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	fmt.Println(userNames)

	userName, err := genorm.
		Pluck(orm.User(), user.NameExpr).
		OrderBy(genorm.Desc, user.Name).
		Where(genorm.EqLit(user.IDExpr, userID)).
		Get(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	fmt.Println(userName)

	userCountVals, err := genorm.
		Pluck(orm.User(), genorm.Count(user.IDExpr, false)).
		GroupBy(user.Name).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	for _, userCountVal := range userCountVals {
		if userCount, ok := userCountVal.Val(); ok {
			fmt.Println(userCount)
		}
	}

	userIDColumn := orm.MessageUserParseExpr(user.ID)
	messageUserIDColumn := orm.MessageUserParseExpr(message.UserID)
	messageCreatedAtColumn := orm.MessageUserParseExpr(message.CreatedAt)
	messageUserValues, err := genorm.
		Select(orm.Message().
			User().Join(genorm.Eq(userIDColumn, messageUserIDColumn))).
		Where(genorm.And(
			genorm.EqLit(userIDColumn, userID),
			genorm.GeqLit(messageCreatedAtColumn, genorm.Wrap(time.Now().Add(-time.Hour))),
		)).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select message user: %w", err)
	}
	fmt.Println(messageUserValues)

	affectedRows, err = genorm.
		Update(orm.Message()).
		Set(
			genorm.AssignLit(message.Content, genorm.Wrap("hello world")),
			genorm.AssignLit(message.CreatedAt, genorm.Wrap(time.Now())),
		).
		Where(genorm.EqLit(message.IDExpr, messageID1)).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	fmt.Println(affectedRows)

	affectedRows, err = genorm.
		Delete(orm.Message()).
		Where(genorm.EqLit(message.UserIDExpr, userID)).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	fmt.Println(affectedRows)

	return nil
}
