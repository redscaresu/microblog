package microblog

import (
	"database/sql"
)

type PostgresStore struct {
	DB *sql.DB
}

func (p *PostgresStore) GetAll() ([]string, error) {
	// horrible SQL goes here
	// handle errors, etc
	return posts, nil
}

func (p *PostgresStore) Create(post string) error {
	return nil
}

// func (ps *PostgresStore) Retrieve(ID string) (Widget, error) {
// 	w := Widget{}
// 	ctx := context.Background()
// 	row := ps.DB.QueryRowContext(ctx, "SELECT id, name FROM widgets WHERE id = ?", ID)
// 	err := row.Scan(&w.ID, &w.Name)
// 	if err != nil {
// 		return Widget{}, err
// 	}
// 	return w, nil
// }
