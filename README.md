
## What is this?

This is an attempt to learn Go. The intention is to make a little app that can get song data and store it in a database. Once I have that data I can start dong cool things. 

## What's the thinking?

For now the goal is having all the data. 

Phase 1. 
- Connect to Genius/Spotify/wherever and get info. 
- Scrape the lyrics from the relevant pages. 
- Store this all in a database structure that makes sense (artists, albums, songs, etc)

Phase 2. 
- Show this data somehow. Initially through a simple server approach that can deliver json or whatever. 
- Using htmx would be great! 
- ????


## What do I need to run this? 
1. Go needs to be installed. 
1. A .env file with API keys. (Will elaborate later)

## How do I run this? 

`go run main.go`