//Contains structs for the database

package main

type stat struct {
	ID     string
	Game   string
	Hours  float64
	Ignore bool
}

type icon struct {
	Name string
	URL  string
	Hash string
}

type setting struct {
	ID              string
	GraphType       string
	MentionForStats bool
}
