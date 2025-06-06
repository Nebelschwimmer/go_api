package bookservice_test

import (
	"api/train/mapper"
	"api/train/models/dto"
	"api/train/models/entities"
	bookservice "api/train/services/book"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListBooks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "title", "release_year", "summary", "price", "cover",
		"author_id", "firstname", "lastname", "birthday",
	}).
		AddRow(1, "Tom Sawyer", 1976, "An adventure book", 12.2, "cover1.jpg", 1, "Mark", "Twain", "1835-11-30").
		AddRow(2, "The Red Hat", 1986, "A fairy tale book", 14.2, "cover2.jpg", 2, "Charles", "Perrault", "1628-01-12")

	query := `
		SELECT b.id, b.title, b.release_year, b.summary, b.price, b.cover,
			a.id, a.firstname, a.lastname, a.birthday
		FROM book b
		LEFT JOIN author a ON b.author_id = a.id
	`

	mock.ExpectQuery(query).WillReturnRows(rows)

	books, err := bookservice.List(db)
	assert.NoError(t, err)
	assert.Len(t, books, 2)
	assert.Equal(t, "Tom Sawyer", books[0].Title)

}

func TestFindBookWithMockDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()

	bookId := 1

	rows := sqlmock.NewRows([]string{
		"id", "title", "release_year", "summary", "price", "cover",
		"id", "firstname", "lastname", "birthday",
	}).
		AddRow(1, "Tom Sawyer", 1976, "An adventure book", 12.2, "cover1.jpg", 1, "Mark", "Twain", "1835-11-30")

	query := `
        SELECT b.id, b.title, b.release_year, b.summary, b.price, b.cover,
            a.id, a.firstname, a.lastname, a.birthday
        FROM book b
        LEFT JOIN author a ON b.author_id = a.id
        WHERE b.id = \$1
    `
	mock.ExpectQuery(query).WithArgs(bookId).WillReturnRows(rows)

	book, err := bookservice.Find(bookId, db)
	assert.NoError(t, err)
	cover := "cover1.jpg"
	expected := entities.Book{
		ID:          1,
		Title:       "Tom Sawyer",
		ReleaseYear: 1976,
		Summary:     "An adventure book",
		Price:       12.2,
		Cover:       &cover,
		Author: entities.Author{
			ID:        1,
			Firstname: "Mark",
			Lastname:  "Twain",
			Birthday:  "1835-11-30",
		},
	}
	mapped := mapper.MapToBookResponse(&expected)

	assert.Equal(t, mapped, book)
}

func TestCreateBookWithMockDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockAuthorFind(mock, 1)

	dto := dto.BookDto{
		Title:       "Tom Sawyer",
		ReleaseYear: 1923,
		Summary:     "An adventure book",
		Price:       12.2,
		AuthorID:    1,
	}

	mock.ExpectQuery("INSERT INTO book \\(title, release_year, summary, price, author_id\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\) RETURNING id").
		WithArgs(dto.Title, dto.ReleaseYear, dto.Summary, dto.Price, dto.AuthorID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	createdBook, err := bookservice.Create(dto, db)
	assert.NoError(t, err)
	assert.Equal(t, 1, createdBook)
}

func TestUpdateBookWithcMockDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()
	bookId := 1

	mockAuthorFind(mock, 1)

	dto := dto.BookDto{
		Title:       "Tom Sawyer",
		ReleaseYear: 1923,
		Summary:     "An adventure book",
		Price:       12.2,
		AuthorID:    1,
	}

	mock.ExpectExec("UPDATE book SET title = \\$1, release_year = \\$2, summary = \\$3, price = \\$4, author_id = \\$5 WHERE id = \\$6").
		WithArgs(dto.Title, dto.ReleaseYear, dto.Summary, dto.Price, dto.AuthorID, bookId).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = bookservice.Update(bookId, dto, db)
	assert.NoError(t, err)
}

func TestDeleteBookWithMockDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock db: %v", err)
	}
	defer db.Close()
	bookId := 1

	mock.ExpectExec("DELETE FROM book WHERE id = \\$1").
		WithArgs(bookId).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = bookservice.Delete(bookId, db)
	assert.NoError(t, err)
}

func mockAuthorFind(mock sqlmock.Sqlmock, id int) {
	mock.ExpectQuery("SELECT id, firstname, lastname, birthday FROM author WHERE id = \\$1").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "firstname", "lastname", "birthday"}).
			AddRow(id, "Mark", "Twain", time.Now()))
}
