package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/cjp2600/protoc-gen-structify/example/case_two/db"
)

type userServiceStr struct {
	store db.UserCRUDOperations // use minimal interface
}

func newUserService(store db.UserCRUDOperations) *userServiceStr {
	return &userServiceStr{store: store}
}

func (u *userServiceStr) Create(ctx context.Context, user *db.User, relations db.Option) (*string, error) {
	return u.store.Create(ctx, user, db.WithRelations())
}

func _main() {
	connection, err := db.Open(db.Dsn("case_two", "", "", 3))
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	store := db.NewBlogStorages(connection)
	userStorage := store.GetUserStorage()
	userService := newUserService(userStorage)
	{
		if err := store.CreateTables(ctx); err != nil {
			panic(err)
		}
	}

	// Create Transaction Manager for store
	err = store.TxManager().ExecFuncWithTx(ctx, func(ctx context.Context) error {

		// Create user with all fields and relations
		id, err := userService.Create(ctx, &db.User{
			Name:  "John",
			Age:   21,
			Email: "example@mail.com",
			Device: &db.Device{
				Name:  "Samsung",
				Value: "foo",
			},
			Settings: &db.Setting{
				Name:  "is_active",
				Value: "true",
			},
			Addresses: []*db.Address{
				{
					City:   "Moscow",
					Street: "Lenina",
					State:  12,
				},
				{
					City:   "Berlin",
					Street: "Dzerzhinskogo",
					State:  112,
				},
			},
			NotificationSettings: &db.UserNotificationSetting{
				RegistrationEmail: true,
				OrderEmail:        false,
			},
			Phones: db.NewPhonesField([]string{"+7 999 999 99 99", "+7 999 999 99 98"}),
			Balls:  db.NewBallsField([]int32{1, 2, 3, 4, 5}),
			Numrs: db.NewNumrsField([]db.UserNumr{
				{
					Street: "Lenina",
					State:  12,
					City:   "Moscow",
				},
				{
					City:   "Berlin",
					Street: "Dzerzhinskogo",
					State:  112,
				},
			}),
			Comments: db.NewCommentsField([]db.UserComment{
				{
					Name: "John",
					Meta: &db.CommentMeta{
						Ip:      "10.0.0.1",
						Browser: "Opera",
						Os:      "Windows",
					},
				},
			}),
		}, db.WithRelations())
		if err != nil {
			return err
		}

		fmt.Printf("User id: %s \n", *id)
		return nil
	})
	if err != nil {
		if errors.Is(err, db.ErrRowAlreadyExist) {
			fmt.Println("user already exists")
			return
		}
	}
}
