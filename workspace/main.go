package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/mazrean/genorm"
	orm "github.com/mazrean/genorm-workspace/workspace/genorm"
	"github.com/mazrean/genorm-workspace/workspace/genorm/message"
	"github.com/mazrean/genorm-workspace/workspace/genorm/user"
	"github.com/mazrean/genorm-workspace/workspace/types"
)

func main() {
	dbEnv, ok := os.LookupEnv("DB")
	if !ok {
		panic("DB is not set")
	}

	db, err := openDatabase(dbEnv)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = migration(dbEnv, db)
	if err != nil {
		panic(err)
	}

	err = runQuery(db)
	if err != nil {
		panic(err)
	}
}

type wrappedDB struct {
	*sql.DB
}

func (db *wrappedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	log.Printf("query: %s\n", query)
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *wrappedDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	log.Printf("query row: %s\n", query)
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *wrappedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	log.Printf("exec: %s\n", query)
	return db.DB.ExecContext(ctx, query, args...)
}

func openDatabase(dbEnv string) (*wrappedDB, error) {
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

	var (
		db  *sql.DB
		err error
	)
	switch dbEnv {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FTokyo&charset=utf8mb4",
			user,
			pass,
			host,
			port,
			database,
		)

		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host,
			port,
			user,
			pass,
			database,
		)

		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown database: %s", dbEnv)
	}

	return &wrappedDB{
		DB: db,
	}, nil
}

func migration(dbEnv string, db *wrappedDB) error {
	var (
		driver database.Driver
		err    error
	)
	switch dbEnv {
	case "mysql":
		driver, err = mysql.WithInstance(db.DB, &mysql.Config{})
		if err != nil {
			return fmt.Errorf("failed to create driver: %w", err)
		}
	case "postgres":
		driver, err = postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			return fmt.Errorf("failed to create driver: %w", err)
		}
	default:
		return fmt.Errorf("unknown database: %s", dbEnv)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://migration/%s", dbEnv),
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

func runQuery(db *wrappedDB) error {
	// INSERT INTO `users` (`id`, `name`, `created_at`) VALUES ({{uuid.New()}}, "name", {{time.Now()}})
	affectedRows, err := genorm.
		Insert(orm.User()).
		Values(&orm.UserTable{
			ID:        types.UserID(uuid.New()),
			Name:      genorm.Wrap("name"),
			CreatedAt: genorm.Wrap(time.Now()),
		}).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}
	log.Printf("affected rows: %d\n", affectedRows)

	// INSERT INTO `users` (`id`, `name`) VALUES ({{uuid.New()}}, "name")
	affectedRows, err = genorm.
		Insert(orm.User()).
		Fields(user.ID, user.Name).
		Values(&orm.UserTable{
			ID:   types.UserID(uuid.New()),
			Name: genorm.Wrap("name"),
		}).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}
	log.Printf("affected rows: %d\n", affectedRows)

	// INSERT INTO `users` (`id`, `name`, `created_at`) VALUES ({{uuid.New()}}, "name", {{time.Now()}})
	affectedRows, err = genorm.
		Insert(orm.User()).
		Values(&orm.UserTable{
			ID:        types.UserID(uuid.New()),
			Name:      genorm.Wrap("name"),
			CreatedAt: genorm.Wrap(time.Now()),
		}).
		DoCtx(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}
	log.Printf("affected rows: %d\n", affectedRows)

	userID := types.UserID(uuid.New())
	affectedRows, err = genorm.
		Insert(orm.User()).
		Values(&orm.UserTable{
			ID:        userID,
			Name:      genorm.Wrap("user"),
			CreatedAt: genorm.Wrap(time.Now()),
		}).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	log.Printf("affected rows: %d\nuserID: %+v\n", affectedRows, userID)

	messageID1 := types.MessageID(uuid.New())
	messageID2 := types.MessageID(uuid.New())
	affectedRows, err = genorm.
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
	log.Printf("affected rows: %d\nmessageID1: %+v\nmessageID2: %+v\n", affectedRows, messageID1, messageID2)

	// SELECT `id`, `name`, `created_at` FROM `users`
	// userValues: []orm.UserTable
	userValues, err := genorm.
		Select(orm.User()).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `id`, `name`, `created_at` FROM `users`
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		GetAllCtx(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `id`, `name`, `created_at` FROM `users` WHERE `id` = {{uuid.New()}}
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Where(genorm.EqLit(user.IDExpr, types.UserID(uuid.New()))).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `name`, `created_at` FROM `users`
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Fields(user.Name, user.CreatedAt).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `id`, `name`, `created_at` FROM `users` LIMIT 1
	// userValue: orm.UserTable
	userValue, err := genorm.
		Select(orm.User()).
		Get(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValue: %+v\n", userValue)

	// SELECT `id`, `name`, `created_at` FROM `users` LIMIT 1
	// userValue: orm.UserTable
	userValue, err = genorm.
		Select(orm.User()).
		GetCtx(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValue: %+v\n", userValue)

	// SELECT `id`, `name`, `created_at` FROM `users` ORDER BY `created_at` DESC
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		OrderBy(genorm.Desc, user.CreatedAt).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `id`, `name`, `created_at` FROM `users` ORDER BY `created_at` DESC, `id` ASC
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		OrderBy(genorm.Desc, user.CreatedAt).
		OrderBy(genorm.Asc, user.ID).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT DISTINCT `id`, `name`, `created_at` FROM `users`
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Distinct().
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `id`, `name`, `created_at` FROM `users` LIMIT 5 OFFSET 3
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Limit(5).
		Offset(3).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `name` FROM `users` GROUP BY `name` HAVING COUNT(`id`) > 10
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Fields(user.Name).
		GroupBy(user.Name).
		Having(genorm.GtLit(genorm.Count(user.IDExpr, false), genorm.Wrap(int64(10)))).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `id`, `name`, `created_at` FROM `users` FOR UPDATE
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Lock(genorm.ForUpdate).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)

	// SELECT `id` FROM `users`
	// userIDs: []uuid.UUID
	userIDs, err := genorm.
		Pluck(orm.User(), user.IDExpr).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userIDs: %+v\n", userIDs)

	// SELECT `id` FROM `users` LIMIT 1
	// userID: uuid.UUID
	userID, err = genorm.
		Pluck(orm.User(), user.IDExpr).
		Get(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userID: %+v\n", userID)

	// SELECT `users`.`name`, `messages`.`content` FROM `users` INNER JOIN `messages` ON `users`.`id` = `messages`.`user_id`
	// messageUserValues: []orm.MessageUserTable
	userIDExpr := orm.MessageUserParseExpr(user.ID)
	userName := orm.MessageUserParse(user.Name)
	messageUserID := orm.MessageUserParseExpr(message.UserID)
	messageContent := orm.MessageUserParse(message.Content)
	messageUserValues, err := genorm.
		Select(orm.User().
			Message().Join(genorm.Eq(userIDExpr, messageUserID))).
		Fields(userName, messageContent).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("messageUserValues: %+v\n", messageUserValues)

	// UPDATE `users` SET `name`="name"
	affectedRows, err = genorm.
		Update(orm.User()).
		Set(
			genorm.AssignLit(user.Name, genorm.Wrap("name")),
		).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	log.Printf("affectedRows: %+v\n", affectedRows)

	// UPDATE `users` SET `name`="name"
	affectedRows, err = genorm.
		Update(orm.User()).
		Set(
			genorm.AssignLit(user.Name, genorm.Wrap("name")),
		).
		DoCtx(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	log.Printf("affectedRows: %+v\n", affectedRows)

	// UPDATE `users` SET `name`="name" WHERE `id`={{uuid.New()}}
	affectedRows, err = genorm.
		Update(orm.User()).
		Set(
			genorm.AssignLit(user.Name, genorm.Wrap("name")),
		).
		Where(genorm.EqLit(user.IDExpr, types.UserID(uuid.New()))).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	log.Printf("affectedRows: %+v\n", affectedRows)

	// UPDATE `users` SET `name`="name" ORDER BY `created_at` LIMIT 1
	affectedRows, err = genorm.
		Update(orm.User()).
		Set(
			genorm.AssignLit(user.Name, genorm.Wrap("name")),
		).
		OrderBy(genorm.Desc, user.CreatedAt).
		Limit(1).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	log.Printf("affectedRows: %+v\n", affectedRows)

	// DELETE FROM `users`
	affectedRows, err = genorm.
		Delete(orm.User()).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	log.Printf("affectedRows: %+v\n", affectedRows)

	// DELETE FROM `users`
	affectedRows, err = genorm.
		Delete(orm.User()).
		DoCtx(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	log.Printf("affectedRows: %+v\n", affectedRows)

	// DELETE FROM `users` WHERE `id`={{uuid.New()}}
	affectedRows, err = genorm.
		Delete(orm.User()).
		Where(genorm.EqLit(user.IDExpr, types.UserID(uuid.New()))).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	log.Printf("affectedRows: %+v\n", affectedRows)

	// DELETE FROM `users` ORDER BY `created_at` LIMIT 1
	affectedRows, err = genorm.
		Delete(orm.User()).
		OrderBy(genorm.Desc, user.CreatedAt).
		Limit(1).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	log.Printf("affectedRows: %+v\n", affectedRows)

	tx, _ := db.Begin()
	// SELECT `id`, `name`, `created_at` FROM `users` FOR UPDATE
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Lock(genorm.ForUpdate).
		GetAll(tx)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	log.Printf("userValues: %+v\n", userValues)
	_ = tx.Commit()

	// SELECT * FROM `messages` WHERE `messages`.`id`=`messages`.`user_id`
	/* compile error
	messageValues, err := genorm.
			Select(orm.Message()).
			Where(genorm.Eq(message.IDExpr, message.UserIDExpr)).
			GetAll(db)
	*/

	return nil
}
