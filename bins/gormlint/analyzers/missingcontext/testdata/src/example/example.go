package example

import "gorm.io/gorm"

type svc struct {
	db *gorm.DB
}

type User struct {
	ID   string
	Name string
}

func (s *svc) badNoContext() {
	var user User
	s.db.Where(User{Name: "alice"}).First(&user) // want `GORM query chain missing WithContext\(\)`
	_ = user
}

func (s *svc) badNoContextFind() {
	var users []User
	s.db.Where(User{Name: "alice"}).Find(&users) // want `GORM query chain missing WithContext\(\)`
	_ = users
}

func (s *svc) goodWithContext() {
	var user User
	s.db.WithContext(nil).Where(User{Name: "alice"}).First(&user)
	_ = user
}

func (u *User) AfterQuery(tx *gorm.DB) error {
	var related User
	tx.Where(User{Name: "related"}).First(&related)
	return nil
}

func goodLocalDB(db *gorm.DB) {
	var user User
	db.Where(User{Name: "alice"}).First(&user)
}
