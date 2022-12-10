package cmd

import (
	"context"
	"fmt"
	"github.com/9d4/semaphore/store"
	"github.com/go-redis/redis/v9"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	v "github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type bootData struct {
	db  *gorm.DB
	rdb *redis.Client
}

type (
	cobraFunc func(cmd *cobra.Command, args []string)
	bootFunc  func(cmd *cobra.Command, args []string, passData *bootData)
)

func boot(fn bootFunc) cobraFunc {
	return func(cmd *cobra.Command, args []string) {
		// connect db and something else here
		data := &bootData{}

		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			v.GetString("dbhost"),
			v.GetInt("dbport"),
			v.GetString("dbuser"),
			v.GetString("dbpasswd"),
			v.GetString("dbname"),
		)
		db, err := gorm.Open(postgres.Open(dsn))
		if err != nil {
			jww.FATAL.Fatal(err)
		}
		data.db = db

		rdb := redis.NewClient(&redis.Options{
			Addr:     v.GetString("REDIS_ADDR"),
			Username: v.GetString("REDIS_USERNAME"),
			Password: v.GetString("REDIS_PASSWORD"),
		})

		if err = rdb.Ping(context.Background()).Err(); err != nil {
			jww.FATAL.Fatal(err)
		}

		// build store
		data.rdb = rdb

		// auto migrate
		jww.INFO.Print("Auto Migrating...")
		store.MigrateAll(db)
		jww.INFO.Print("done.")

		fn(cmd, args, data)
	}
}
