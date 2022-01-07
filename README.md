# Distributed-Services

This project is essentially a result of my attempt to understand the concepts of Distributed Sytems, and implement them
with Golang to consequently build a fully-fletched distributed service. Since, the learning curve itself is madly steep, I have separated the project developement in the stages listed below. However, the main aim is to build
a distributed service with it's very own storage handling, networking over a client and server, and a way to
distribute the server instances. At the end, if possible, I plan to deploy the service with Kubernetes to the cloud.

At this point in time ( 8th Janauary 2022), the first step has been successfully completed and tested. The second step is
expected to be completed by the end of January.

The stages were decided in this order to reflect the content structure of the book Distributed Services with Go, written by Travis Jeffery.
As the book proceeds with the concepts, I have tried to simultaneously learn and build the different components of the service alongside.
Finally, the stages are as follows :

 ### Building the project's storage layer, a web server to faciliate JSON over HTTP, and a custom made log libray
  - Develop the JSON over HTTP commit log service
  - Setup protobufs, and ways to aumatically generate the data structures based on the protobuf message structures
  - Building a commit log library that will essentially be the log for the entire service, to store and lookup data
    - The commit log library has the following structure:
        - A Component that allows appending and reading records from the log by provisioning indpendent structures and methods to faciliate
          - Store file handling for record entries
          - Index file handling for index entries of the corresponding records
        - A Component that combines the store and file handling components to provision a Segment file handling module to coordinate operations across record and index files.
        - Lasty, the final component ties all the components above, specially the Segment module, to create the final Log handling package for the enitre libaray.
    - All the files for the log library can be found under the internal/log directory.
   
### Creating the service over a network
  - Setting up gRPC, define the client server APIs in protobuf together with builing the client and server
  - Securing the service with authentication of the server with SSL/TLS, to encrypy/decrypt the data exchanged by authenticating requests with accress tokens.
  - Making service observable by addings logs, metrics and tracing
  
### Distribute - making the service distributed
  - Building discovery into service to make server instances aware of each other
  - Adding Raft consensus to coordinate the efforts of our servers, and turn them into a cluster
  - Putting discovery into out gRPC clients, so that discover and connect to servers with client side load balancing.



## Development 
 The entire development of the project is dependent on my learning curve, and ability to grasp the concepts of distirbued services. Since, the 
 project is entirely for educational purposes, it is hard to predict a possible timeline. However, by the end of this month, the entire project can
 be expected to be completed.

## Disclaimer
As stated above, the project is being built by closely following the content and concepts outlined in the book 
Distributed Services with Go by Travis Jeffery. Hence, it's being developed purely for my own exploration and learning of distribued services.
Howvever, anyone willing to contribue is welcomed.

## Author
Hamza Yusuff - Email: hbyusuff@uwaterloo.ca

