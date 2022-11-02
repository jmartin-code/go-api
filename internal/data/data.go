package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const dbTimeout = time.Second * 3

var db *sql.DB

func New(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		User:  User{},
		Token: Token{},
	}
}

type Models struct {
	User  User
	Token Token
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Token     Token     `json:"token"`
}

func (u *User) GetAllUsers() ([]*User, error) {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, created_at, updated_at from users order by last_name`

	rows, err := db.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}
	return users, nil
}

func (u *User) GetUserByEmail(email string) (*User, error) {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, created_at, updated_at where email = $1`

	row := db.QueryRowContext(ctx, query, email)

	var user User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *User) GetUserById(id int) (*User, error) {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, created_at, updated_at from users where id = $1`

	row := db.QueryRowContext(ctx, query, id)

	var user User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *User) UpdateUser(id int) error {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `update users set
			email = $1,
			first_name = $2,
			last_name = $3,
			updated_at = $4
			where id = $5
	`

	_, err := db.ExecContext(ctx, query,
		u.Email,
		u.FirstName,
		u.LastName,
		time.Now(),
		u.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) AddUser(user User) (int, error) {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	//create has password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)

	if err != nil {
		return 0, err
	}

	var userId int

	query := `insert into users (email, first_name, last_name, password, created_at, updated_at) values ($1, $2, $3, $4, $5, $6) returning id`

	err = db.QueryRowContext(ctx, query,
		user.Email,
		user.FirstName,
		user.LastName,
		hashedPassword,
		time.Now(),
		time.Now(),
	).Scan(&userId)

	if err != nil {
		println("error inside addUser")
		return 0, err
	}

	return userId, nil
}

func (u *User) ResetUserPassword(password string) error {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	//create has password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	if err != nil {
		return err
	}

	query := `update user set password = $1 where id = $2`

	_, err = db.ExecContext(ctx, query, hashedPassword, u.ID)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) UserPasswordMatch(plainPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPassword))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			//invalid password
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (u *User) DeleteUser() error {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `delete from users where id = $1`

	_, err := db.ExecContext(ctx, query, u.ID)

	if err != nil {
		return err
	}

	return nil
}

type Token struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	TokenHash []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Expiry    time.Time `json:"expiry"`
}

func (t *Token) GetUserByToken(plainText string) (*Token, error) {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, user_id, email, token, token_hash, created_at, updated_at, expiry from tokens where token = $1`

	var token Token

	row := db.QueryRowContext(ctx, query, plainText)

	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.Email,
		&token.Token,
		&token.TokenHash,
		&token.CreatedAt,
		&token.UpdatedAt,
		&token.Expiry,
	)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (t *Token) GetUserForToken(token Token) (*User, error) {
	// if it takes longer than 3 seconds, cancel
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, created_at, updated_at from users where id = $1`

	row := db.QueryRowContext(ctx, query, token.UserID)

	var user User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// generate token
func (t *Token) GenerateToken(userID int, ttl time.Duration) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Token = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Token))
	token.TokenHash = hash[:]

	return token, nil
}

// Authenthicate
func (t *Token) AuthenticateToken(r *http.Request) (*User, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("no authorization header")
	}

	headersParts := strings.Split(authorizationHeader, " ")
	if len(headersParts) != 2 || headersParts[0] != "Bearer" {
		return nil, errors.New("no valid header")
	}

	token := headersParts[1]

	if len(token) != 26 {
		return nil, errors.New("token size is not valid")
	}

	tk, err := t.GetUserByToken(token)
	if err != nil {
		return nil, errors.New("not matching token found")
	}

	if tk.Expiry.Before(time.Now()) {
		return nil, errors.New("expired token")
	}

	user, err := t.GetUserForToken(*tk)
	if err != nil {
		return nil, errors.New("not matching user found")
	}

	return user, nil
}

func (t *Token) DeleteByToken(plain string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `delete from tokens where token = $1`

	_, err := db.ExecContext(ctx, query, plain)

	if err != nil {
		return err
	}

	return nil
}

func (t *Token) InsertToken(token Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `delete from tokens where user_id = $1`

	_, err := db.ExecContext(ctx, query, token.UserID)

	if err != nil {
		return err
	}

	query = `insert into tokens (user_id, email, token, token_hash, created_at, updated_at, expiry) values ($1, $2, $3, $4, $5, $6, $7)`

	_, err = db.ExecContext(ctx, query,
		token.UserID,
		token.Email,
		token.Token,
		token.TokenHash,
		time.Now(),
		time.Now(),
		token.Expiry,
	)

	if err != nil {
		return err
	}

	return nil
}

func (t *Token) ValidToken(plain string) (bool, error) {
	token, err := t.GetUserByToken(plain)

	if err != nil {
		return false, errors.New("no matching found")
	}

	_, err = t.GetUserForToken(*token)
	if err != nil {
		return false, errors.New("not matching user found")
	}

	if token.Expiry.Before(time.Now()) {
		return false, errors.New("expired token")
	}

	return true, nil
}
