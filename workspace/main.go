package main

import (
	"context"
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
	"github.com/mazrean/genorm-workspace/workspace/types"
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
	fmt.Println(affectedRows, userID)

	messageID1 := types.MessageID(uuid.New())
	messageID2 := types.MessageID(uuid.New())
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

	// SELECT `id`, `name`, `created_at` FROM `users`
	// userValues: []orm.UserTable
	userValues, err := genorm.
		Select(orm.User()).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}

	// SELECT `id`, `name`, `created_at` FROM `users`
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		GetAllCtx(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}

	// SELECT `id`, `name`, `created_at` FROM `users` WHERE `id` = {{uuid.New()}}
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Where(genorm.EqLit(user.IDExpr, types.UserID(uuid.New()))).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}

	// SELECT `name`, `created_at` FROM `users`
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Fields(user.Name, user.CreatedAt).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	fmt.Println(userValues)

	// SELECT `id`, `name`, `created_at` FROM `users` LIMIT 1
	// userValue: orm.UserTable
	userValue, err := genorm.
		Select(orm.User()).
		Get(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}

	// SELECT `id`, `name`, `created_at` FROM `users` LIMIT 1
	// userValue: orm.UserTable
	userValue, err = genorm.
		Select(orm.User()).
		GetCtx(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	fmt.Println(userValue)

	// SELECT `id`, `name`, `created_at` FROM `users` ORDER BY `created_at` DESC
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		OrderBy(genorm.Desc, user.CreatedAt).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}

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

	// SELECT DISTINCT `id`, `name`, `created_at` FROM `users`
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Distinct().
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}

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

	// SELECT `id`, `name`, `created_at` FROM `users` FOR UPDATE
	// userValues: []orm.UserTable
	userValues, err = genorm.
		Select(orm.User()).
		Lock(genorm.ForUpdate).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}

	// SELECT `id` FROM `users`
	// userIDs: []uuid.UUID
	userIDs, err := genorm.
		Pluck(orm.User(), user.IDExpr).
		GetAll(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}
	fmt.Println(userIDs)

	// SELECT `id` FROM `users` LIMIT 1
	// userID: uuid.UUID
	userID, err = genorm.
		Pluck(orm.User(), user.IDExpr).
		Get(db)
	if err != nil {
		return fmt.Errorf("failed to select user: %w", err)
	}

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
	fmt.Println(messageUserValues)

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
	fmt.Println(affectedRows)

	// DELETE FROM `users`
	affectedRows, err = genorm.
		Delete(orm.User()).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// DELETE FROM `users`
	affectedRows, err = genorm.
		Delete(orm.User()).
		DoCtx(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// DELETE FROM `users` WHERE `id`={{uuid.New()}}
	affectedRows, err = genorm.
		Delete(orm.User()).
		Where(genorm.EqLit(user.IDExpr, types.UserID(uuid.New()))).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// DELETE FROM `users` ORDER BY `created_at` LIMIT 1
	affectedRows, err = genorm.
		Delete(orm.User()).
		OrderBy(genorm.Desc, user.CreatedAt).
		Limit(1).
		Do(db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	fmt.Println(affectedRows)

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
	fmt.Println(userValues)
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
