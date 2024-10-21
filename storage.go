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
