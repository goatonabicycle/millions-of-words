
## What is this?

This is me learning Go and htmx. Ultimately, this will be a mechanism to explore the lyrics of albums in a couple of interesting ways.

## What's the thinking?

- Get all the data.
  - Have a mechanism of getting the information from Bandcamp. 
  - Only bandcamp? The music I care about is currently mostly on Bandcamp. This had a mechanism
    where I'd get all the albums from Spotify and then get the lyrics from Genius. That worked but it's only 
    consistent if you're looking for "popular" music. Genius seems to be hip hop focused. 
- Clean the data
  - Sometimes people have [This person's part:] mixed into the actual lyrics.
- Show the data
  - Have a way of counting and sorting word usage. 
  - Sentiment stuff.
  - I might be able to create some interesting visualisations based on word length


- Turns out hosting this is a bit of a pain. I could go full serverless with a database.

## What do I need to run this? 
Just Go at the moment.

## How do I run this? 

`go run main.go`

## Deploy: 

Deploy: `flyctl deploy` 
