# Design decisions

## Architectural characteristics

I've defined scalability (elastisity in future), availability and maintainability as the main architecture characteristics.

### Scalability

To cope with high load we have to distribute the load over multiple nodes, and do scale out by increasing nodes count dynamically, based on load.

If the load unevenly distibuted over the time of the day (that is very likely for a game, because of in working hours the load will lower) to save the costs the computation powers should be elastic and shrink when load is decreased.

### Availability

Also during rolled updates and disaters we can loose some nodes and that shouldn't affect user experience.

### Maintainability

The changes and new features should be introduced fast, that is why project must have understandable structure with the short learining curve.

## Decisions

### Commands

For serving the workloads the cluster of actors was chosen (Proto.Actor framework). It provides all of the characteristics described above.
The cluster runs a top of Kubernetes, and use it for cluster configuration and discovery.

In additional actor systems provide high throughtput, because active actors are located in memory, and there is no need to get them from database for serving the request.

Actor state persistance (not implemented in the assesment) is achieved by saving events into DB. Because events usuially much smaller than full state that allow update DB much faster, that also affects high throughtput and DB load.

If the node where actor gone offline, the actor is recreated on another available node from the state stored in DB.

### Queries

The proportion queries/commands usually is 5:1. That is why the query side serves querires directly from Redis, without interacting with actor's cluster.

### Maintainability

Project stucture organized according Hexagonal Architecture with folders structured by different layers of application:

- domain
- application
- adapters

# Deploy into kubernetes

The prerequisite for installation is the pre-installed Helm.

Run `install.sh` in the root of the repo and application will be deployed into `leaderboard` namespace of the current k8s context.

Game parameters (competition duration, size, etc) could be confugured by setting up ENV variables of the container. For more info see `./src/application/config.go`
