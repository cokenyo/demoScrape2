package main

type game struct {
	//winnerID         int
	coreID           string
	mapNum           int
	winnerClanName   string
	result           string
	rounds           []*round
	potentialRound   *round
	teams            map[string]*team
	flags            flag
	mapName          string
	tickRate         int
	tickLength       int
	roundsToWin      int //30 or 16
	totalPlayerStats map[uint64]*playerStats
	ctPlayerStats    map[uint64]*playerStats
	tPlayerStats     map[uint64]*playerStats
	totalTeamStats   map[string]*teamStats
	playerOrder      []uint64
	teamOrder        []string
	totalRounds      int
	totalWPAlog      []*wpalog
}

type flag struct {
	//all our sentinals and shit
	hasGameStarted            bool
	isGameLive                bool
	isGameOver                bool
	inRound                   bool
	prePlant                  bool
	postPlant                 bool
	postWinCon                bool
	roundIntegrityStart       int
	roundIntegrityEnd         int
	roundIntegrityEndOfficial int

	//for the round (gets reset on a new round) maybe should be in a new struct
	tAlive            int
	ctAlive           int
	tMoney            bool
	tClutchVal        int
	ctClutchVal       int
	tClutchSteam      uint64
	ctClutchSteam     uint64
	openingKill       bool
	lastTickProcessed int
	ticksProcessed    int
	didRoundEndFire   bool
	roundStartedAt    int
	haveInitRound     bool
}

type team struct {
	//id    int //meaningless?
	name          string
	score         int
	scoreAdjusted int
}

type teamStats struct {
	winPoints      float64
	impactPoints   float64
	tWinPoints     float64
	ctWinPoints    float64
	tImpactPoints  float64
	ctImpactPoints float64
	_4v5w          int
	_4v5s          int
	_5v4w          int
	_5v4s          int
	pistols        int
	pistolsW       int
	saves          int
	clutches       int
	traded         int
	fass           int
	ef             int
	ud             int
	util           int
	ctR            int
	ctRW           int
	tR             int
	tRW            int
	deaths         int

	//kinda garbo
	normalizer int
}

type round struct {
	//round value
	roundNum            int8
	startingTick        int
	endingTick          int
	playerStats         map[uint64]*playerStats
	teamStats           map[string]*teamStats
	initTerroristCount  int
	initCTerroristCount int
	winnerClanName      string
	//winnerID            int //this is the unique ID which should not change BUT IT DOES
	winnerENUM         int //this effectively represents the side that won: 2 (T) or 3 (CT)
	integrityCheck     bool
	planter            uint64
	defuser            uint64
	endDueToBombEvent  bool
	winTeamDmg         int
	serverNormalizer   int
	serverImpactPoints float64
	knifeRound         bool

	WPAlog        []*wpalog
	bombStartTick int
}

type wpalog struct {
	round               int
	tick                int
	clock               int
	planted             int
	ctAlive             int
	tAlive              int
	ctEquipVal          int
	tEquipVal           int
	ctFlashes           int
	ctSmokes            int
	ctMolys             int
	ctFrags             int
	tFlashes            int
	tSmokes             int
	tMolys              int
	tFrags              int
	closestCTDisttoBomb int
	kits                int
	ctArmor             int
	tArmor              int
	winner              int
}

type playerStats struct {
	name    string
	steamID uint64
	isBot   bool
	//teamID  int
	teamENUM     int
	teamClanName string
	side         int
	rounds       int
	//playerPoints float32
	//teamPoints float32
	damage              int
	kills               uint8
	assists             uint8
	deaths              uint8
	deathTick           int
	deathPlacement      float64
	ticksAlive          int
	trades              int
	traded              int
	ok                  int
	ol                  int
	cl_1                int
	cl_2                int
	cl_3                int
	cl_4                int
	cl_5                int
	_2k                 int
	_3k                 int
	_4k                 int
	_5k                 int
	nadeDmg             int
	infernoDmg          int
	utilDmg             int
	ef                  int
	fAss                int
	enemyFlashTime      float64
	hs                  int
	kastRounds          float64
	saves               int
	entries             int
	killPoints          float64
	impactPoints        float64
	winPoints           float64
	awpKills            int
	RF                  int
	RA                  int
	nadesThrown         int
	firesThrown         int
	flashThrown         int
	smokeThrown         int
	damageTaken         int
	suppRounds          int
	suppDamage          int
	lurkerBlips         int
	distanceToTeammates int
	lurkRounds          int
	wlp                 float64
	mip                 float64
	rws                 float64 //round win shares
	eac                 int     //effective assist contributions

	rwk int //rounds with kills

	//derived
	utilThrown   int
	atd          int
	kast         float64
	killPointAvg float64
	iiwr         float64
	adr          float64
	drDiff       float64
	kR           float64
	tr           float64 //trade ratio
	impactRating float64
	rating       float64

	//side specific
	tDamage               int
	ctDamage              int
	tImpactPoints         float64
	tWinPoints            float64
	tOK                   int
	tOL                   int
	ctImpactPoints        float64
	ctWinPoints           float64
	ctOK                  int
	ctOL                  int
	tKills                uint8
	tDeaths               uint8
	tKAST                 float64
	tKASTRounds           float64
	tADR                  float64
	ctKills               uint8
	ctDeaths              uint8
	ctKAST                float64
	ctKASTRounds          float64
	ctADR                 float64
	tTeamsWinPoints       float64
	ctTeamsWinPoints      float64
	tWinPointsNormalizer  int
	ctWinPointsNormalizer int
	tRounds               int
	ctRounds              int
	ctRating              float64
	ctImpactRating        float64
	tRating               float64
	tImpactRating         float64
	tADP                  float64
	ctADP                 float64

	tRF   int
	ctAWP int

	//kinda garbo
	teamsWinPoints      float64
	winPointsNormalizer int

	//"flags"
	health             int
	tradeList          map[uint64]int
	mostRecentFlasher  uint64
	mostRecentFlashVal float64
	damageList         map[uint64]int
}

type Accolades struct {
	awp        int
	deagle     int
	knife      int
	dinks      int
	blindKills int
	bombPlants int
	jumps      int
	teamDMG    int
	selfDMG    int
	ping       int
	pingPoints int
	//footsteps         int //unnecessary processing?
	bombTaps          int
	killsThroughSmoke int
	penetrations      int
	noScopes          int
	midairKills       int
	crouchedKills     int
	bombzoneKills     int
	killsWhileMoving  int
	mostMoneySpent    int
	mostShotsOnLegs   int
	shotsFired        int
	ak                int
	m4                int
	pistol            int
	scout             int
}
