# go-challonge

Golang API client for retrieving (and potentially updating) tournaments in  [Challonge](http://challonge.com/).

## Install

    $ go get github.com/FlowingSPDG/go-challonge

## Usage

### Load and initialize

    package main
    import "github.com/FlowingSPDG/go-challonge"
    
    fun main() {
        client := challonge.New("challonge-user", "challonge-key")
    }


### Tournaments

Retrieve tournament

    t, err := client.NewTournamentRequest("tournament").Get()
    
    if err != nil {
        // invalid tournament name, unable to reach host etc.
        log.Fatal("unable to retrieve tournament ", err)
    }
    
Retrieve tournament with matches and participants

    t, err := client.NewTournamentRequest("tournament").
            WithMatches().
            WithParticipants().
            Get()
            
To re-fetch a tournament

    newTournament, err := oldTournament.Update().Get()
    
Create a new tournament. Requires name, url, subdomain (can be an empty string), whether to be open or not and tournament type (defaults to single for empty string).

    t, err := client.CreateTournament("name", "url", "subdomain", true, "single")
    
### Matches

Get a list of all open matches:

    matches := t.GetOpenMatches()
    
Get a specific match:

    match := t.GetMatch(id)

### Participants

Add participant to tournament. Misc-field is an API specific field which can be used to identify users.

    p, err := t.AddParticipant("name", "misc")
    
Remove a participant

    // by name
    err := t.RemoveParticipant("name")
    // by id
    err := t.RemoveParticipantById(id)
