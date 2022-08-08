package products

import (
	"database/sql"
	"log"

	"github.com/nictes1/storage-implementation/internal/domain"
)

type Repository interface {
	Store(domain.Product) (domain.Product, error)
	GetOne(id int) (domain.Product, error)
	Update(id int, name, productType string, count int, price float64) (domain.Product, error)
	GetAll() ([]domain.Product, error)
	Delete(id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Store(product domain.Product) (domain.Product, error) { // se inicializa la base
	stmt, err := r.db.Prepare("INSERT INTO products(name, type, count, price) VALUES( ?, ?, ?, ? )") // se prepara la sentencia SQL a ejecutar
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close() // se cierra la sentencia al terminar. Si quedan abiertas se genera consumos de memoria
	var result sql.Result
	result, err = stmt.Exec(product.Name, product.Type, product.Count, product.Price) // retorna un sql.Result y un error
	if err != nil {
		return domain.Product{}, err
	}
	insertedId, _ := result.LastInsertId() // del sql.Resul devuelto en la ejecucion obtenemos el Id insertado
	product.ID = int(insertedId)
	return product, nil
}

func (r *repository) GetOne(id int) (domain.Product, error) {
	var product domain.Product

	rows, err := r.db.Query("SELECT id, name, type, count, price FROM products WHERE id = ?", id)
	if err != nil {
		return domain.Product{}, err
	}
	for rows.Next() {
		if err := rows.Scan(&product.ID, &product.Name, &product.Type, &product.Count, &product.Price); err != nil {
			return domain.Product{}, err
		}
	}
	return product, nil
}

func (r *repository) Update(id int, name, productType string, count int, price float64) (domain.Product, error) {
	stmt, err := r.db.Prepare("UPDATE products SET name = ?, type = ?, count = ?, price = ? WHERE id = ?") // se prepara la sentencia SQL a ejecutar
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close() // se cierra la sentencia al terminar. Si quedan abiertas se genera consumos de memoria
	product := domain.Product{ID: id, Name: name, Type: productType, Count: count, Price: price}
	_, err = stmt.Exec(name, productType, count, price, id)
	if err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

const (
	GetAllProducts = "SELECT * FROM products"
)

func (r *repository) GetAll() ([]domain.Product, error) {
	var products []domain.Product
	rows, err := r.db.Query(GetAllProducts)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// se recorren todas las filas
	for rows.Next() {
		// por cada fila se obtiene un objeto del tipo Product
		var product domain.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Type, &product.Count, &product.Price); err != nil {
			log.Fatal(err)
			return nil, err
		}
		//se añade el objeto obtenido al slice products
		products = append(products, product)
	}
	return products, nil
}

func (r *repository) Delete(id int) error { // se inicializa la base
	stmt, err := r.db.Prepare("DELETE FROM products WHERE id = ?") // se prepara la sentencia SQL a ejecutar
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()     // se cierra la sentencia al terminar. Si quedan abiertas se genera consumos de memoria
	_, err = stmt.Exec(id) // retorna un sql.Result y un error
	if err != nil {
		return err
	}
	return nil
}
