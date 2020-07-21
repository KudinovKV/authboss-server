package database

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/go-pg/pg"
	zl "github.com/rs/zerolog/log"
	"github.com/volatiletech/authboss/v3"
)

// User struct for authboss and Postgres
type User struct {
	ID int `sql:"id,pk"`

	// Non-authboss related field
	Name string `sql:"name"`
	Role string `sql:"role"`

	// Auth
	Email    string `sql:"email"`
	Password string `sql:"password"`

	// Confirm
	ConfirmSelector string `sql:"confirmselector"`
	ConfirmVerifier string `sql:"confirmverifier"`
	Confirmed       bool   `sql:"confirmed"`

	// Lock
	AttemptCount int       `sql:"attemptcount"`
	LastAttempt  time.Time `sql:"lastattempt"`
	Locked       time.Time `sql:"locked"`

	// Recover
	RecoverSelector    string    `sql:"recoverselector"`
	RecoverVerifier    string    `sql:"recoververifier"`
	RecoverTokenExpiry time.Time `sql:"recovertokenexpiry"`

	// Remember is in another table
}

// This pattern is useful in real code to ensure that
// we've got the right interfaces implemented.
var (
	db           *pg.DB
	assertUser   = &User{}
	assertStorer = &Storer{}

	_ authboss.User            = assertUser
	_ authboss.AuthableUser    = assertUser
	_ authboss.ConfirmableUser = assertUser
	_ authboss.LockableUser    = assertUser
	_ authboss.RecoverableUser = assertUser
	_ authboss.ArbitraryUser   = assertUser

	_ authboss.CreatingServerStorer    = assertStorer
	_ authboss.ConfirmingServerStorer  = assertStorer
	_ authboss.RecoveringServerStorer  = assertStorer
	_ authboss.RememberingServerStorer = assertStorer
)

// PutID into user
func (u *User) PutPID(pid string) {
	u.Email = pid
}

// PutID into user
func (u *User) PutRole(role string) {
	u.Role = role
}

// PutEmail into user
func (u *User) PutEmail(email string) { u.Email = email }

// PutPassword into user
func (u *User) PutPassword(password string) { u.Password = password }

// PutConfirmed into user
func (u *User) PutConfirmed(confirmed bool) { u.Confirmed = confirmed }

// PutConfirmSelector into user
func (u *User) PutConfirmSelector(confirmSelector string) { u.ConfirmSelector = confirmSelector }

// PutConfirmVerifier into user
func (u *User) PutConfirmVerifier(confirmVerifier string) { u.ConfirmVerifier = confirmVerifier }

// PutLocked into user
func (u *User) PutLocked(locked time.Time) { u.Locked = locked }

// PutAttemptCount into user
func (u *User) PutAttemptCount(attempts int) { u.AttemptCount = attempts }

// PutLastAttempt into user
func (u *User) PutLastAttempt(last time.Time) { u.LastAttempt = last }

// PutRecoverSelector into user
func (u *User) PutRecoverSelector(token string) { u.RecoverSelector = token }

// PutRecoverVerifier into user
func (u *User) PutRecoverVerifier(token string) { u.RecoverVerifier = token }

// PutRecoverExpiry into user
func (u *User) PutRecoverExpiry(expiry time.Time) { u.RecoverTokenExpiry = expiry }

// PutArbitrary into user
func (u *User) PutArbitrary(values map[string]string) {
	if n, ok := values["name"]; ok {
		u.Name = n
	}
	if r, ok := values["role"]; ok {
		u.Role = r
	}
}

// GetID into user
func (u User) GetPID() string { return u.Email }

// GetRole into user
func (u User) GetRole() string { return u.Role }

// GetEmail from user
func (u User) GetEmail() string { return u.Email }

// GetPassword from user
func (u User) GetPassword() string { return u.Password }

// GetConfirmed from user
func (u User) GetConfirmed() bool { return u.Confirmed }

// GetConfirmSelector from user
func (u User) GetConfirmSelector() string { return u.ConfirmSelector }

// GetConfirmVerifier from user
func (u User) GetConfirmVerifier() string { return u.ConfirmVerifier }

// GetLocked from user
func (u User) GetLocked() time.Time { return u.Locked }

// GetAttemptCount from user
func (u User) GetAttemptCount() int { return u.AttemptCount }

// GetLastAttempt from user
func (u User) GetLastAttempt() time.Time { return u.LastAttempt }

// GetRecoverSelector from user
func (u User) GetRecoverSelector() string { return u.RecoverSelector }

// GetRecoverVerifier from user
func (u User) GetRecoverVerifier() string { return u.RecoverVerifier }

// GetRecoverExpiry from user
func (u User) GetRecoverExpiry() time.Time { return u.RecoverTokenExpiry }

// GetArbitrary from user
func (u User) GetArbitrary() map[string]string {
	return map[string]string{
		"name": u.Name,
	}
}

type Storer struct {
	Users  map[string]User
	Tokens map[string][]string
}

// NewStorer return pointer to copy of Storer struct
func NewStorer(d *pg.DB) *Storer {
	db = d
	dbUsers := []User{}
	err := db.Model(&dbUsers).Select()
	if err != nil {
		zl.Warn().Err(err)
	}
	users := map[string]User{}
	for _, user := range dbUsers {
		users[user.Email] = user
	}
	return &Storer{
		Users:  users,
		Tokens: make(map[string][]string)}
}

// New user creation
func (s Storer) New(_ context.Context) authboss.User {
	return &User{}
}

// Create the user
func (s Storer) Create(_ context.Context, user authboss.User) error {
	u := user.(*User)
	u.ID = len(s.Users) + 1
	if _, ok := s.Users[u.Email]; ok {
		return authboss.ErrUserFound
	}
	zl.Debug().Msgf("Created new user:", u.Name)
	s.Users[u.Email] = *u
	return nil
}

// Save the user
func (s Storer) Save(_ context.Context, user authboss.User) error {
	// TODO: save user to database
	abUser := user.(*User)
	s.Users[abUser.Email] = *abUser
	_, err := db.Model(abUser).Insert()
	if err != nil {
		_, err := db.Model(abUser).Where("ID = ?", abUser.ID).Update()
		zl.Debug().Msgf("Update user %v", abUser.Name)
		return err
	}
	zl.Debug().Msgf("Save user %v", abUser.Name)
	return nil
}

// Load the user
func (s Storer) Load(_ context.Context, key string) (authboss.User, error) {
	// TODO: load user from database
	abUser, ok := s.Users[key]
	if !ok {
		return nil, authboss.ErrUserNotFound
	}
	zl.Debug().Msgf("Load user %v", abUser.Name)
	return &abUser, nil
}

// LoadByConfirmSelector looks a user up by confirmation token
func (s Storer) LoadByConfirmSelector(_ context.Context, selector string) (user authboss.ConfirmableUser, err error) {
	for _, v := range s.Users {
		if v.ConfirmSelector == selector {
			zl.Debug().Msgf("Loaded user by confirm selector:", selector, v.Name)
			return &v, nil
		}
	}

	return nil, authboss.ErrUserNotFound
}

// LoadByRecoverSelector looks a user up by confirmation selector
func (s Storer) LoadByRecoverSelector(_ context.Context, selector string) (user authboss.RecoverableUser, err error) {
	for _, v := range s.Users {
		if v.RecoverSelector == selector {
			zl.Debug().Msgf("Loaded user by recover selector:", selector, v.Name)
			return &v, nil
		}
	}

	return nil, authboss.ErrUserNotFound
}

// AddRememberToken to a user
func (s Storer) AddRememberToken(_ context.Context, pid, token string) error {
	s.Tokens[pid] = append(s.Tokens[pid], token)
	zl.Debug().Msgf("Adding rm token to %s: %s\n", pid, token)
	spew.Dump(s.Tokens)
	return nil
}

// DelRememberTokens removes all tokens for the given pid
func (s Storer) DelRememberTokens(_ context.Context, pid string) error {
	delete(s.Tokens, pid)
	zl.Debug().Msgf("Deleting rm tokens from:", pid)
	spew.Dump(s.Tokens)
	return nil
}

// UseRememberToken finds the pid-token pair and deletes it.
// If the token could not be found return ErrTokenNotFound
func (s Storer) UseRememberToken(_ context.Context, pid, token string) error {
	tokens, ok := s.Tokens[pid]
	if !ok {
		zl.Debug().Msgf("Failed to find rm tokens for:", pid)
		return authboss.ErrTokenNotFound
	}

	for i, tok := range tokens {
		if tok == token {
			tokens[len(tokens)-1] = tokens[i]
			s.Tokens[pid] = tokens[:len(tokens)-1]
			zl.Debug().Msgf("Used remember for %s: %s\n", pid, token)
			return nil
		}
	}

	return authboss.ErrTokenNotFound
}

// InitDb connect to database and create table
func InitDb(toConnect string, createOrNot bool) (*pg.DB, error) {
	pgOpt, err := pg.ParseURL(toConnect)
	if err != nil {
		return nil, err
	}
	pgdb := pg.Connect(pgOpt)
	if pgdb == nil {
		zl.Fatal().Err(err).
			Msg("Can't connect to database")
	}
	if createOrNot {
		_, err = pgdb.Exec(`
		CREATE TABLE Users 
		(
			ID integer PRIMARY KEY,
			Name varchar(30) NOT NULL,
			Email varchar(30) NOT NULL,
			Password varchar(129) NOT NULL,
			Role varchar(30) NOT NULL,
			confirmselector varchar(129),
			confirmverifier varchar(129),
			confirmed bool,
			attemptcount integer,
			lastattempt timestamp,
			locked timestamp,
			recoverselector varchar(129),
			recoververifier varchar(129),
			recovertokenexpiry timestamp
		);`)
		if err != nil {
			zl.Fatal().Err(err).
				Msg("Can't create table")
		}
	}

	return pgdb, nil
}
