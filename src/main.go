package main

func main() {
	NewApp(
		&Storage{
			ConnectionInfo: ConnectionInfo{
				driver:   "postgres",
				username: "root",
				password: "secret",
				host:     "localhost",
				port:     5432,
				database: "movies",
			},
		},
	).Run(8000)
}
