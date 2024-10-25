
# Task Description
Implement a REST API for a leaderboard service. API endpoints and what the client expects to receive from them are defined below.

The leaderboard service groups players to competitions of 10 players, balancing the matchmaking waiting time and grouping participants with others that are close to their level (and as a bonus, if you have time, country). Both data points for matchmaking are given in a JSON document described below.
Once enough players have joined a competition (or 30 seconds of waiting has passed), the competition starts. The competition runs for 60 minutes, during which players can submit points. The points are added incrementally to players’ total points in this competition.

The REST API allows retrieving the player’s current competition by player ID, or any past or current competition by competition ID. In both endpoints, the player IDs and their total score are returned, ordered by the points. When two players are tied, they are sorted alphabetically by player ID.
After the competition has ended, the player can no longer submit points to it and needs to join another leaderboard competition to compete again. In the new competion, the player starts from zero points.
## API Endpoints
`POST /leaderboard/join?player_id=<string ID>`
A player calls this endpoint to start matchmaking into a leaderboard competition.
If a player has already joined a competition and that competition has not ended, they cannot join a new one until the current competition ends.
Success responses:
If a matching competition is found right away:
```
{
  "leaderboard_id": "<string uuid>",
  "ends_at": <timestamp>,
}
```
If a competition fitting the player is not found right away, the player needs to wait in a matchmaking queue, and the service returns `HTTP 202 (Accepted)`.
If the player is already in an active competition and cannot join, the service returns `HTTP 409 (conflict)`.
When picking a player’s competition group, you can use the following player data structure (assuming it’s coming from a database, but you don’t need to write the code 
for retrieving it):
```
{
  "level": <int>,
  "country_code": "<string>",

}
```
Note: keep concurrency in mind, e.g. what if two players join a competition at the same time?
Also, matchmaking shouldn’t take more than 30 seconds. So, when designing your matchmaking algorithm, balance the time it takes to find the match with how close the players are to each others’ levels, in the end, placing the player in the closest matching group.
If you have time, it would be nice to see a similar implementation based on country code, but the requirement is to just handle player level as input data.
`GET /leaderboard/player/<player_id>`
The player can only participate in one leaderboard at a time, so this endpoint returns the information about the player’s current leaderboard, or the last one they joined, if the current one has finished.
This can be used e.g. to check if the matchmaking was completed.
Returns:
```
{
  "leaderboard_id": "<string id>",
  "ends_at": <timestamp>,
  "leaderboard": [
    {
      "player_id": "<string id>",
      "score": <int>
}, {
      "player_id": "<string id>",
      "score": <int>
    },
... ]
}
```
If the player is not in a leaderboard, return an empty JSON object.
`GET /leaderboard/<leaderboardID>`
Returns the information about a specific leaderboard competition. This can be useful for e.g. returning historical leaderboard information.
Returns:
```
{
  "leaderboard_id": "<string id>",
  "ends_at": <timestamp>,
  "leaderboard": [
    {
      "player_id": "<string id>",
      "score": <int>
}, {
      "player_id": "<string id>",
      "score": <int>
    },
... ]
}
```
If a leaderboard with the given ID doesn't exist, returns `HTTP 404 (Not found)`.
`POST /leaderboard/score`
Submits a score to the player's current leaderboard. The submitted score is added to the player's total score without server side validation.
JSON body:
```
{
  "player_id": "<string id>",
  "score": <int>
}
```
If the player has not joined a competition or the latest competition is over, the service returns `HTTP 409 (conflict)`.
In a success case, the service returns HTTP 200 (OK) without a body.

# Design decisions

In contrast with a regular stateless web application, the game application is mostly stateful:

- high command rate: it's not possible to store and reload state from the DB for each command from the player
- related data should be placed on the same servers to minimize network delays

One of the possible ways to deal with such kind stateful applications with respect to arch charecteristics listed below, is an actor model.

## Architectural characteristics

I've defined scalability (elasticity in the future), availability, and maintainability as the main architectural characteristics.

### Scalability

To cope with high load we have to distribute the load over multiple nodes and do scale out by increasing node count dynamically, based on load.

If the load is unevenly distributed over the time of the day (that is very likely for a game, because of in working hours the load will lower) to save the costs the computation powers should be elastic and shrink when the load is decreased.

### Availability

Also during rolled updates and disasters, we can loose some nodes and that shouldn't affect user experience.

### Maintainability

The changes and new features should be introduced fast, that is why the project must have understandable structure with the short learining curve.

## Decisions

### Commands

For serving the workloads the cluster of actors was chosen ([Proto.Actor](https://proto.actor/) framework) and its [Grains](https://proto.actor/docs/cluster/) . It provides all of the characteristics described above.
The cluster runs a top of Kubernetes, and uses it for cluster configuration and discovery.

In additional actor, systems provide high throughtput, because active actors are located in memory, and there is no need to get them from database for serving the request.

Actor state persistance (not implemented in the assesment) is achieved by saving events into DB. Because events usuially much smaller than full state that allow update DB much faster, that also affects high throughtput and DB load.

If the node where actor gone offline, the actor is recreated on another available node from the state stored in DB.

### Queries

The proportion queries/commands usually is 5:1. That is why the query side serves querires directly from Redis, without interacting with actor's cluster.

### Maintainability

Project structure organized according to Hexagonal Architecture with folders structured by different layers of application:

- domain
- application
- adapters

# Deploy into Kubernetes

The prerequisite for installation is the pre-installed Helm.

- run the command `kubectl get storageclasses` and copy the value of your cluster storage class name.
- update `stroageClass` value in `leaderboard-helm/values.yaml` with your cluster storage class.
- run `install.sh` in the root of the repo and application will be deployed into `leaderboard` namespace of the current k8s context.

# Testing

Game parameters (competition duration, size, etc) could be configured by setting up values in `game` section of `leaderboard-helm/values.yaml`

For easier testing competition time is reduced from 1 hour to 200 seconds.

Competition will be started there are at least 2 players in the same level bracket

Level brackets are 0-10, 11-20 and 21-30

There are 30 players with ids from 1 to 30, the player's ID corresponds to the player level, e.g. player with id=5 has level=5

Any other players have 1 level.
