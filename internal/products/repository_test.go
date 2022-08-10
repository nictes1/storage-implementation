package products

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/DATA-DOG/go-txdb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/nictes1/storage-implementation/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	//mock.ExpectPrepare("INSERT INTO products(name, type, count, price) VALUES( ?, ?, ?, ? )")
	mock.ExpectPrepare("INSERT INTO products")
	mock.ExpectExec("INSERT INTO products").WillReturnResult(sqlmock.NewResult(1, 1))
	productId := 1

	repo := NewRepo(db)
	product := domain.Product{
		ID:    productId,
		Name:  "Teclado",
		Type:  "Periferico",
		Count: 2,
		Price: 340.5,
	}

	p, err := repo.Store(product)
	assert.NoError(t, err)
	assert.NotZero(t, p)
	assert.Equal(t, product.ID, p.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryGetAllOK(t *testing.T) {
}

func TestGetAllConflict(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM products")).WillReturnError(sql.ErrConnDone)

	repo := NewRepo(db)
	//ctx := context.TODO()
	result, err := repo.GetAll()

	assert.Equal(t, sql.ErrConnDone, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOneOk(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "type", "count", "price"})
	product := domain.Product{
		ID:    000455,
		Name:  "Teclado",
		Type:  "Periferico",
		Count: 2,
		Price: 340.5,
	}
	rows.AddRow(product.ID, product.Name, product.Type, product.Count, product.Price)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM products WHERE id=?")).WithArgs(product.ID).WillReturnRows(rows)

	repo := NewRepo(db)
	result, err := repo.GetOne(product.ID)
	assert.Equal(t, product, result)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

}

func TestDeleteOK(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	id := 1

	mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM products WHERE id=?"))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM products WHERE id=?")).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewRepo(db)
	err = repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM products WHERE id=?")).WillReturnError(sql.ErrNoRows)
	_, err = repo.GetOne(id)
	assert.ErrorContains(t, sql.ErrNoRows, err.Error())
}

func TestDeleteErrorExec(t *testing.T) {
}

//Test store con TXDB
func TestRepositoryStoreTXDB(t *testing.T) { //Realizamos mock sobre la transaccion de la base de datos. Luego hace un rollback

	txdb.Register("txdb", "mysql", "root@tcp(localhost:3306)/storage")

	//sql.Open recibe el driver de base de datos y un string de conexion
	db, err := sql.Open("txdb", uuid.New().String())
	assert.NoError(t, err)

	repo := NewRepo(db) //Generamos nuestro repository
	product := domain.Product{
		Name:  "Iphone",
		Type:  "Tecnologia",
		Count: 987,
		Price: 1200,
	}
	p, err := repo.Store(product) //consulta el repo.
	assert.NoError(t, err)
	assert.NotZero(t, p)
}

//Extra
func TestRepositoryGetWithTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	productId := 1
	columns := []string{"id", "name", "type", "count", "price"}
	rows := sqlmock.NewRows(columns)
	rows.AddRow(productId, "remera", "indumentaria", 3, 1500)
	mock.ExpectQuery("select id, name, type, count, price").WillDelayFor(10 * time.Second).WillReturnRows(rows)
	repo := NewRepo(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = repo.GetOneWithContext(ctx, productId)

	assert.Error(t, err)
}
