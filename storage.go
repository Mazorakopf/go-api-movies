package main

type Movie struct {
	ID       string    `json:"id"`
	Isbn     string    `json:"isbn"`
	Title    string    `json:"title"`
	Director *Director `json:"director"`
}

type Director struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

var movies = []Movie{
	{ID: "1", Isbn: "438227", Title: "Movie One", Director: &Director{FirstName: "Uladzislau", LastName: "Kleshchanka"}},
	{ID: "2", Isbn: "454551", Title: "Movie Two", Director: &Director{FirstName: "John", LastName: "Doe"}},
}

var users = []map[string]string{
	{"name": "admin", "password": "$2a$10$miwrWXNyiF7Qiv6ir9YTueSulDDrJfjj1w2r1dLpAmaOj/TglYyKG"},
}
