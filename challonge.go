package challonge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	API_VERSION = "v1"
	tournaments = "tournaments"
	STATE_OPEN  = "open"
	STATE_ALL   = "all"
)

var client *Client
var debug = false

type tournament Tournament

type Client struct {
	baseUrl string
	key     string
	version string
	user    string
}

type APIResponse struct {
	Tournament  *Tournament  `json:"tournament"`
	Participant *Participant `json:"participant"`
	Match       Match        `json:"match"`

	Errors []string `json:"errors"`
}

type Tournament struct {
	Name              string     `json:"name"`
	Id                int        `json:"id"`
	Url               string     `json:"url"`
	FullUrl           string     `json:"full_challonge_url"`
	State             string     `json:"state"`
	SubDomain         string     `json:"subdomain"`
	ParticipantsCount int        `json:"participants_count"`
	StartedAt         *time.Time `json:"started_at,mitempty"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
	Type              string     `json:"tournament_type"`
	Description       string     `json:"description"`
	GameName          string     `json:"game_name"`
	Progress          int        `json:"progress_meter"`

	SubUrl string `json:"sub_url"`

	ParticipantItems []*ParticipantItem `json:"participants,omitempty"`
	MatchItems       []*MatchItem       `json:"matches,omitempty"`

	Participants []*Participant `json:"resolved_participants"`
	Matches      []*Match       `json:"resolved_matches"`
}

type Participant struct {
	Id         int    `json:"id"`
	Name       string `json:"display_name"`
	Misc       string `json:"misc"`
	Seed       int    `json:"seed"`
	FinalRank  int    `json:"final_rank"`
	Wins       int
	Losses     int
	TotalScore int
}

type Match struct {
	Id                   int    `json:"id"`
	Identifier           string `json:"identifier"`
	State                string `json:"state"`
	Round                int    `json:"round"`
	PlayerOneId          int    `json:"player1_id"`
	PlayerOnePrereqMatch *int   `json:"player1_prereq_match_id"`
	PlayerTwoId          int    `json:"player2_id"`
	PlayerTwoPrereqMatch *int   `json:"player2_prereq_match_id"`
	PlayerOneScore       int
	PlayerTwoScore       int
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`

	WinnerId int `json:"winner_id"`

	PlayerOne *Participant
	PlayerTwo *Participant
	Winner    *Participant

	Scores string `json:"scores_csv"`
}

/** items to flatten json structure */
type TournamentItem struct {
	Tournament Tournament `json:"tournament"`
}

type ParticipantItem struct {
	Participant Participant `json:"participant"`
}

type MatchItem struct {
	Match *Match `json:"match"`
}

type TournamentRequest struct {
	client *Client
	Id     string
	Params map[string]string
}

func (c *Client) Print() {
	log.Print(c.key)
}

func New(user string, key string) *Client {
	client = &Client{user: user, version: API_VERSION, key: key}
	return client
}

func (c *Client) Debug() {
	debug = true
}

func (c *Client) buildUrl(route string, v url.Values) string {
	url := fmt.Sprintf("https://%s:%s@api.challonge.com/%s/%s.json", c.user, c.key, c.version, route)
	if v != nil {
		url += "?" + v.Encode()
	}

	return url
}

func params(p map[string]string) *url.Values {
	values := url.Values{}
	for k, v := range p {
		values.Add(k, v)
	}
	return &values
}

func (r *APIResponse) hasErrors() bool {
	if debug {
		log.Printf("response had errors: %q", r.Errors)
	}
	return len(r.Errors) > 0
}

func (r *APIResponse) getTournament() *Tournament {
	return r.Tournament.resolveRelations()
}

type GetTournamentsResponse struct {
	Tournament *Tournament `json:"tournament"`
	//Errors []string `json:"errors"`
}

// GetTournaments Get tournaments that belongs to your account.
func (c *Client) GetTournaments(state string, rtype string, subdomain string) ([]*Tournament, error) {
	v := *params(map[string]string{})
	v = *params(map[string]string{
		"state": state, // all, pending, in_progress, ended
		"type":  rtype, // single elimination, double elimination, round robin, swiss
	})
	if subdomain != "" {
		v.Set("subdomain", subdomain)
	}
	url := client.buildUrl("tournaments", v)
	response := []GetTournamentsResponse{}
	doGet(url, &response)
	tournaments := make([]*Tournament, 0, len(response))
	for i := 0; i < len(response); i++ {
		tournaments = append(tournaments, response[i].Tournament)
	}
	return tournaments, nil
}

func (c *Client) NewTournamentRequest(id string) *TournamentRequest {
	return &TournamentRequest{client: c, Id: id, Params: make(map[string]string, 0)}
}

func (r *TournamentRequest) WithParticipants() *TournamentRequest {
	r.Params["include_participants"] = "1"
	return r
}

func (r *TournamentRequest) WithMatches() *TournamentRequest {
	r.Params["include_matches"] = "1"
	return r
}

func (t *Tournament) Update() *TournamentRequest {
	return client.NewTournamentRequest(t.SubUrl)
}

func (r *TournamentRequest) Get() (*Tournament, error) {
	url := r.client.buildUrl("tournaments/"+r.Id, *params(r.Params))
	response := &APIResponse{}
	doGet(url, response)
	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("unable to retrieve tournament: %q", response.Errors[0])
	}
	tournament := response.getTournament()
	tournament.SubUrl = r.Id
	return tournament, nil
}

type CreateTournamentRequest struct {
	Name                            string `json:"name"`
	URL                             string `json:"url"`
	Description                     string `json:"description"`
	RequireScoreAgreement           bool   `json:"require_score_agreement"`
	NotifyUserWhenMatchesOpen       bool   `json:"notify_users_when_matches_open"`
	OpenSignup                      bool   `json:"open_signup"`
	NotifyUserWhenTheTournamentEnds bool   `json:"notify_users_when_the_tournament_ends"`
	QuickAdvance                    bool   `json:"quick_advance"`
	HoldThirdPlaceMatch             bool   `json:"hold_third_place_match"`
	// Double Elimination or Round Robin related?
	/*
		PTsForGameWin float32 `json:"pts_for_game_win"`
		PTsForGameTie float32 `json:"pts_for_game_tie"`
		PTsForMatchWin float32 `json:"pts_for_match_win"`
		PTsForMatchTie float32 `json:"pts_for_match_tie"`
		PTsForBye float32 `json:"pts_for_bye"`
	*/

	// Swiss Rounds related?
	SwissRounds int `json:"swiss_rounds"`

	Private bool `json:"private"`
	// RankedBy bool `json:"ranked_by"` // "match wins"
	ShowRounds bool `json:"show_rounds"`
	HideForum  bool `json:"hide_forum"`
	// SequentialPairings  bool `json:"sequential_pairings"` // ??
	AcceptAttachments bool `json:"accept_attachments"`

	// Round robin with Group-stage?
	/*
		RoundRobinPTsForGameWin float32 `json:"rr_pts_for_game_win"`
		RoundRobinPTsForGameTie float32 `json:"rr_pts_for_game_tie"`
		RoundRobinPTsForMatchWin float32 `json:"rr_pts_for_match_win"`
		RoundRobinPTsForMatchTie float32 `json:"rr_pts_for_match_tie"`
		RoundRobinPTsForBye float32 `json:"rr_pts_for_bye"`
	*/

	// CreatedByAPI  bool `json:"created_by_api"` // always true?
	// CreditCapped  bool `json:"credit_capped"` // ?
	// Category ??
	HideSeeds bool `json:"hide_seeds"`

	// Prediction related?
	AcceptPredictions               bool `json:"accepting_predictions"` // tournament[voting_enabled] ...?
	PredictionMethod                int  `json:"prediction_method"`
	AnonymousVoting                 bool `json:"anonymous_voting"`
	MaxPredictionsPerUser           int  `json:"max_predictions_per_user"`
	PublicPredictionBeforeStartTime bool `json:"public_predictions_before_start_time"`
	PredictTheLosersBracket         bool `json:"predict_the_losers_bracket"`

	// misc
	SignupCap         int    `json:"signup_cap"`
	GameID            int    `json:"game_id"`            // https://challonge.com/ja/games.json. 194=CSGO.
	GameName          string `json:"game_name"`          // https://challonge.com/ja/games.json. 194=CSGO.
	ParticipantsCount int    `json:"participants_count"` // available participants number?

	// Group stages related
	GroupStagesEabled bool `json:"group_stages_enabled"`

	AllowParticipantMatchReporting bool `json:"allow_participant_match_reporting"`
	Teams                          bool `json:"teams"` // team or persoal participant
	// CheckInDuration check_in_duration // null?

	// Tie breaks?
	/*
			"tie_breaks": [
		        "match wins vs tied",
		        "game wins",
		        "points scored"
		    ],
	*/

	// EventID interface{} `json:"event_id"` // null?
	Ranked bool `json:"ranked"`
	// GrandFinalsModifier interface{} `json:"grand_finals_modifier"` // null?
	// Spam interface{} `json:"spam"` // null?
	// Ham interface{} `json:"ham"` // null...?? WTF?
	RoundRobinIterations int `json:"rr_iterations"`
	// TournamentRegistrationID interface{} `json:"tournament_registration_id"` // null?
	// DonationContestEnabled interface{} `json:"donation_contest_enabled"` // null?
	// MandatoryDonation interface{} `json:"mandatory_donation"` // null?
	/*
			"non_elimination_tournament_data": {
		            "participants_per_match": ""
		        },
	*/
	// AutoAsignStations interface{} `json:"auto_assign_stations"` // null?
	// OnlyStartMatchesWithStations interface{} `json:"only_start_matches_with_stations"` // null?
	RegistrationFee  float32 `json:"registration_fee"`
	RegistrationType string  `json:"registration_type"`
	// SplitParticipants bool `json:"split_participants"` // ?
	ParticipantsSwappable bool `json:"participants_swappable"`
	TeamConvertable       bool `json:"team_convertable"`

	Subdomain string `json:"subdomain"`
}

/** creates a new tournament */
func (c *Client) CreateTournament(name string, subUrl string, domain string, open bool, tType string, desc string) (*Tournament, error) {
	v := *params(map[string]string{
		"tournament[name]":        name,
		"tournament[url]":         subUrl,
		"tournament[open_signup]": "false",
		"tournament[subdomain]":   domain,
		"tournament[description]": desc,
		// tournament[public_sign_up]
		// tournament[organization_id] //
		// tournament[group_stages_enabled] // ?
		// tournament[group_stages_attributes][0][stage_type] // "single elimination", "double elimination", "round robin".
		// tournament[group_stages_attributes][0][rr_iterations] // ?
		// tournament[group_stages_attributes][0][group_size] // Max group team size?
		// tournament[group_stages_attributes][0][participant_count_to_advance_per_group] // ?
		// tournament[group_stages_attributes][0][ranked_by] // ?
		// tournament[group_stages_attributes][0][tie_breaks][] // ?
		// tournament[grand_finals_modifier]
		// tournament[tie_breaks][]// ?
		// tournament[admin_ids_csv] // Admins. comma-separetd
		// tournament_rr_iterations // ??
		// tournament[ranked_by] // ?
		// tournament[start_at] // ?
		// tournament[check_in_duration] // in minute?
		// tournament[voting_enabled] // predictions enabled?
	})
	if tType == "" || tType == "single" {
		v.Add("tournament[tournament_type]", "single elimination")
	} else if tType == "double" {
		v.Add("tournament[tournament_type]", "double elimination")
	} else if tType == "round robin" {
		v.Add("tournament[tournament_type]", "round robin")
	} else if tType == "swiss" {
		v.Add("tournament[tournament_type]", "swiss")
	}
	url := c.buildUrl("tournaments", v)
	response := &APIResponse{}
	doPost(url, response)
	if response.hasErrors() {
		return nil, fmt.Errorf("unable to create tournament: %q", response.Errors[0])
	}
	return response.getTournament(), nil
}

func (t *Tournament) Start() error {
	v := *params(map[string]string{
		"include_participants": "1",
		"include_matches":      "1",
	})
	url := client.buildUrl("tournaments/"+t.GetUrl()+"/start", v)
	response := &APIResponse{}
	doPost(url, response)
	if response.hasErrors() {
		return fmt.Errorf("error starting tournament:  %q", response.Errors[0])
	}
	tournament := response.getTournament()
	if tournament.State == "underway" {
		if debug {
			log.Printf("tournament %q started", tournament.Name)
		}
	} else {
		return fmt.Errorf("tournament has state %q, probably not started", tournament.State)
	}
	t = tournament
	return nil
}

func (t *Tournament) Randomize() error {
	url := client.buildUrl("tournaments/"+t.GetUrl()+"/participants/randomize", nil)
	var response interface{}
	doPost(url, &response)
	if _, ok := response.([]interface{}); ok {
		// fmt.Println("OK?")
		// fmt.Printf("response is []interface{} : %v\n", response.([]interface{}))
	} else if _, ok := response.(map[string]interface{}); ok { //response.(map[string][]string)
		res := response.(map[string]interface{})
		return fmt.Errorf("%v", res["errors"].([]interface{}))
	} else {
		fmt.Println("UNKNOWN JSON")
		fmt.Printf("response is unknown : %v\n", response)
		return fmt.Errorf("Unknown JSON")
	}
	return nil
}

func (t *Tournament) Reset() error {
	v := *params(map[string]string{
		"include_participants": "1",
		"include_matches":      "1",
	})
	url := client.buildUrl("tournaments/"+t.GetUrl()+"/reset", v)
	response := &APIResponse{}
	doPost(url, response)
	if response.hasErrors() {
		return fmt.Errorf("error randomizing participants:  %q", response.Errors[0])
	}
	fmt.Printf("resp : %v\n", response)
	if response == nil {
		return fmt.Errorf("error randomizing participants")
	}
	return nil
}

func (t *Tournament) Destroy() error {
	url := client.buildUrl("tournaments/"+t.GetUrl(), nil)
	response := &APIResponse{}
	doDelete(url, response)
	if response.hasErrors() {
		return fmt.Errorf("error randomizing participants:  %q", response.Errors[0])
	}
	fmt.Printf("resp : %v\n", response)
	if response == nil {
		return fmt.Errorf("error randomizing participants")
	}
	return nil
}

func (t *Tournament) Finalize() error {
	v := *params(map[string]string{
		"include_participants": "1",
		"include_matches":      "1",
	})
	url := client.buildUrl("tournaments/"+t.GetUrl()+"/finalize", v)
	response := &APIResponse{}
	doPost(url, response)
	if response.hasErrors() {
		return fmt.Errorf("error finishing tournament:  %q", response.Errors[0])
	}
	tournament := response.getTournament()
	if tournament.State == "complete" {
		if debug {
			log.Printf("tournament %q completed", tournament.Name)
		}
	} else {
		return fmt.Errorf("tournament has state %q, probably not finished", tournament.State)
	}
	t = tournament
	return nil
}

func (t *Tournament) SubmitMatch(m *Match) (*Match, error) {
	v := *params(map[string]string{
		"match[scores_csv]": fmt.Sprintf("%d-%d", m.PlayerOneScore, m.PlayerTwoScore),
		"match[winner_id]":  fmt.Sprintf("%d", m.WinnerId),
	})
	url := client.buildUrl(fmt.Sprintf("tournaments/%s/matches/%d", t.GetUrl(), m.Id), v)
	response := &APIResponse{}
	doPut(url, response)
	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("%q", response.Errors[0])
	}
	m = &response.Match
	return &response.Match, nil
}

/** adds participant to tournament */
func (t *Tournament) AddParticipant(name string, misc string) (*Participant, error) {
	v := *params(map[string]string{
		"participant[name]": name,
		"participant[misc]": misc,
	})
	url := client.buildUrl("tournaments/"+t.GetUrl()+"/participants", v)
	response := &APIResponse{}
	doPost(url, response)
	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("unable to add participant: %q", response.Errors[0])
	}
	t.Participants = append(t.Participants, response.Participant)
	return response.Participant, nil
}

/** returns "domain-url" or "url" */
func (t *Tournament) GetUrl() string {
	if t.SubDomain != "" {
		return t.SubDomain + "-" + t.Url
	}
	return t.Url
}

/** removes participant from tournament */
func (t *Tournament) RemoveParticipant(name string) error {
	p := t.GetParticipantByName(name)
	if p == nil || p.Id == 0 {
		return fmt.Errorf("participant with name %q not found in tournament", name)
	}
	return t.RemoveParticipantById(p.Id)
}

/** removes participant by id */
func (t *Tournament) RemoveParticipantById(id int) error {
	url := client.buildUrl("tournaments/"+t.GetUrl()+"/participants/"+strconv.Itoa(id), nil)
	response := &APIResponse{}
	doDelete(url, response)
	if len(response.Errors) > 0 {
		return fmt.Errorf("unable to delete participant: %q", response.Errors[0])
	}
	return nil
}

/** returns a participant id based on name */
type cmp func(*Participant) bool

func (t *Tournament) GetParticipant(id int) *Participant {
	return t.getParticipantByCmp(func(p *Participant) bool { return p.Id == id })
}
func (t *Tournament) GetParticipantByName(name string) *Participant {
	return t.getParticipantByCmp(func(p *Participant) bool { return p.Name == name })
}
func (t *Tournament) GetParticipantByMisc(misc string) *Participant {
	return t.getParticipantByCmp(func(p *Participant) bool { return p.Misc == misc })
}

func (t *Tournament) getParticipantByCmp(cmp cmp) *Participant {
	for _, p := range t.Participants {
		if cmp(p) {
			return p
		}
	}
	return nil
}

/** returns all matches for tournament */
func (t *Tournament) GetMatches() []*Match {
	return t.getMatches(STATE_ALL)
}

/** returns all open matches */
func (t *Tournament) GetOpenMatches() []*Match {
	return t.getMatches(STATE_OPEN)
}

/** resolves and returns matches for tournament */
func (t *Tournament) getMatches(state string) []*Match {
	matches := make([]*Match, 0, len(t.Matches))

	for _, m := range t.Matches {
		m.ResolveParticipants(t)
		if state == STATE_ALL {
			matches = append(matches, m)
		} else if m.State == state {
			matches = append(matches, m)
		}
	}
	return matches
}

/** returns match with resolved participants */
func (t *Tournament) GetMatch(id int) *Match {
	for _, match := range t.Matches {
		if match.Id == id {
			match.ResolveParticipants(t)
			return match
		}
	}
	return nil
}

func (t *Tournament) IsCompleted() bool {
	return t.State == "complete" || t.State == "awaiting_review"
}

func (t *Tournament) GetOpenMatchForParticipant(p *Participant) *Match {
	matches := t.GetOpenMatches()
	for _, m := range matches {
		if m.PlayerOneId == p.Id || m.PlayerTwoId == p.Id {
			return m
		}
	}
	return nil
}

func (p *Participant) Lose() {
	p.Losses += 1
}

func (p *Participant) Win() {
	p.Wins += 1
}

func (m *Match) ResolveParticipants(t *Tournament) {
	if m.State != "complete" {
		return
	}
	m.PlayerOne = t.GetParticipant(m.PlayerOneId)
	m.PlayerTwo = t.GetParticipant(m.PlayerTwoId)

	if m.WinnerId == m.PlayerOneId {
		m.PlayerOne.Win()
		m.PlayerTwo.Lose()
	} else if m.WinnerId == m.PlayerTwoId {
		m.PlayerTwo.Win()
		m.PlayerOne.Lose()
	}
	m.PlayerOne.TotalScore += m.PlayerOneScore
	m.PlayerTwo.TotalScore += m.PlayerTwoScore

}

func (t *Tournament) resolveRelations() *Tournament {
	participants := make([]*Participant, 0, len(t.ParticipantItems))
	for _, item := range t.ParticipantItems {
		participants = append(participants, &item.Participant)
	}
	t.Participants = participants
	t.ParticipantItems = nil

	matches := make([]*Match, 0, len(t.MatchItems))
	for _, item := range t.MatchItems {
		match := item.Match
		match.ResolveParticipants(t)
		matches = append(matches, match)
	}
	t.Matches = matches
	t.MatchItems = nil

	return t
}

func DiffMatches(matches1 []*Match, matches2 []*Match) []*Match {
	diff := make([]*Match, 0, len(matches1))

	for i, _ := range matches1 {
		if i >= len(matches2) {
			break
		}
		if matches1[i].State != matches2[i].State {
			diff = append(diff, matches2[i])
		}
	}

	return diff
}

func doGet(url string, v interface{}) {
	if debug {
		log.Print("gets resource on url ", url)
	}
	resp, err := http.Get(url)
	if debug {
		log.Print("got headers ", resp)
	}
	if err != nil {
		log.Fatal("unable to get resource ", err)
	}
	handleResponse(resp, v)
}

func doPost(url string, v interface{}) {
	if debug {
		log.Print("posts resource on url ", url)
	}
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Fatal("unable to get resource ", err)
	}
	handleResponse(resp, v)
}

func doPut(url string, v interface{}) {
	req, err := http.NewRequest("PUT", url, nil)
	log.Print("puts resource on url ", url)
	if err != nil {
		log.Fatal("unable to create put request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("unable to delete", err)
	}
	handleResponse(resp, v)
}

func doDelete(url string, v interface{}) {
	req, err := http.NewRequest("DELETE", url, nil)
	log.Print("deletes resource on url ", url)
	if err != nil {
		log.Fatal("unable to create delete request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("unable to delete", err)
	}
	handleResponse(resp, v)
}

func handleResponse(r *http.Response, v interface{}) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal("unable to read response", err)
	}
	// fmt.Printf("body : %v\n", string(body))
	err = json.Unmarshal(body, v)
	if err != nil {
		//log.Printf("Error unmarshaling json. ERR : %v, BODY : %s\n ", err, string(body))
		log.Printf("Error unmarshaling json. ERR : %s\n ", err)
	}
	if debug {
		log.Print("unmarshaled to ", v)
	}
}

func (t *Tournament) UnmarshalJSON(b []byte) (err error) {
	placeholder := tournament{}
	if err = json.Unmarshal(b, &placeholder); err == nil {
		*t = Tournament(placeholder)
		return
	}
	return
}
