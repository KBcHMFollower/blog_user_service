package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/KBcHMFollower/auth-service/cmd/migrator"
	"github.com/KBcHMFollower/auth-service/internal/domain/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgressStore(connectString string) (*PostgresUserRepository, error){
	db, err := sql.Open("postgres", connectString)

	if err != nil{
		return nil, fmt.Errorf("%w", err)
	}

	return &PostgresUserRepository{db: db}, nil
}

func (r *PostgresUserRepository) Migrate(pathToMigrations string)(error){
	m, err := migrator.New(r.db)
	if err != nil {
		return err
	}

	if err := m.Migrate(pathToMigrations, "postgres"); err != nil {
		return err
	}
	return nil
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context,email string, hashPass []byte) (uuid.UUID, error){
	query := `INSERT INTO users(id, email, pass_hash, created_date, updated_date) VALUES($1, $2, $3, $4, $5)
	  			RETURNING id`

	user := models.NewUserModel(email, hashPass)

	var id uuid.UUID

	err := r.db.QueryRowContext(ctx, query, user.Id, user.Email, user.PassHash, user.CreatedDate, user.UpdatedDate).Scan(&id)
	if err != nil{
		return uuid.New(), fmt.Errorf("%w", err)
	}

	return id, nil
}

func (r *PostgresUserRepository) GetUser(ctx context.Context,email string) (models.User, error){
	query := `SELECT * FROM users
				WHERE email = $1`
	
	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.Id, &user.Email, &user.PassHash, &user.CreatedDate, &user.UpdatedDate)
	if err != nil{
		return models.User{}, err;
	}

	return user, nil
}


