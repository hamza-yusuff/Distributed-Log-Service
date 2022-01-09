# Distributed-Services

This project is essentially a result of my attempt to understand the concepts of Distributed Systems, and implement them
with Golang to consequently build a fully-fletched distributed service. Since, the learning curve itself is madly steep, I have separated the project developement in the stages listed below. However, the main aim is to build
a distributed service with it's very own storage handling, networking over a client and server, and a way to
distribute the server instances. At the end, if possible, I plan to deploy the service with Kubernetes to the cloud.

At this point in time ( 8th Janauary 2022), the first step has been successfully completed and tested. The second step is
expected to be completed by the end of January.

The stages were decided in this order to reflect the content structure of the book Distributed Services with Go, written by Travis Jeffery.
As the book proceeds with the concepts, I have tried to independently and simultaneously learn and build the different components of the service.
Finally, the stages are as follows :

 ### Building the project's storage layer, a web server to faciliate JSON over HTTP, and a custom made log libray
  - Develop the JSON over HTTP commit log service
  - Setup protobufs, and ways to aumatically generate the data structures based on the protobuf message structures
  - Building a commit log library that will essentially be the log for the entire service, to store and lookup data
    - The commit log library has the following structure:
        - A component that allows appending and reading records from the log by provisioning independent structures and methods to faciliate the following
          - Store file handling for record entries
          - Index file handling for index entries of the corresponding records
        - A component that combines the store and file handling components to provision a Segment file handling module to coordinate operations across store and index files.
        - Lasty, the final component ties all the components above, specially the Segment module, to create the final Log handling package for the entire libaray.
    - All the files for the log library can be found under the internal/log directory.
    - To know more about the log library scroll below
   
### Creating the service over a network
  - Setting up gRPC, define the client server APIs in protobuf together with builing the client and server
  - Securing the service with authentication of the server with SSL/TLS, to encrypy/decrypt the data exchanged by authenticating requests with accress tokens.
  - Making service observable by addings logs, metrics and tracing
  
### Distribute - making the service distributed
  - Building discovery into service to make server instances aware of each other
  - Adding Raft consensus to coordinate the efforts of our servers, and turn them into a cluster
  - Putting discovery into out gRPC clients, so that discover and connect to servers with client side load balancing.

## Distributed log library (Overview)

A log basically records what happened and when. It is like a table that always orders the records by the time and the indexes each record by its offset and time created. Logs are split into list of segments to accommodate for the fact that disk spaces are not infinite. When the list of segments grows too big, the old segments are deleted whose data we have already processed. This clean up process usually occurs tin the background or concurrently. Each log contains an active segment, where data is written actively, when the active segment is filled up, the log moves into another segment. Each segment compromises of two files - store and index files. The store files gets record written into, and the index file is where the offset and index value of the record at the store file, is written into.

Logs in distributed system allow ordering changes and distributing data across the nodes in a cluster. More so, the distributed log is said to be a data structure that reflects the problem of consensus. With append-read only logs, each replica in a cluster can read the same input/data from the log for a given instant, and produce the same result. 

This distributed log library is  built to support a replicated, coordinated cluster. It’s done so by adding methods into log.go file that would allow the service to know about the offset range of each log. That way, we would know what nodes have the oldest and newest data, and what nodes are falling behind and need to replicate. Inside the log.go file, there are functions written to read all the segments of a log at It also supports snapshots and restoring of logs when necessary .

It has two separate files index.go and store.go to handle the reading and appending records and indexes into the store and index files respectively. Both .go files have structs for records and indexes, along with corresponding methods to read and append records or indexes. To view the field of the structs defined, you can go to the index.go and store.go files inside the internal/log directory. Also, all the files for the library can be found in the internal/log library. For the offset values in index files, relative offsets as uint32 is used, and not absolute offset values as uint64, to optimise the memory performance. The index files are mostly memory mapped as they are small and has only two data - offset and position of the record inside the store file, and the indexes are appended/read from the sync memory mapped files. This makes the read/append operations faster than what it would have been with the disk involved .

The log has been made to go through a graceful shutdown for the service. Service follows graceful shutdown, and returns the service to a state where it can restart properly and efficiently.  This happens the close method for the index file ( present inside the index.go file ) truncates the persisted file first - by removing the empty spaces between the last record in the index file and the end of file which was there before to compute the maximum possible file size ( Open function did that) . By truncating the persisted file, we remove the empty spaces, and make sure the last entry is the last record appended in the file, and is at the end of file. 

The segment portion (segment.go) wraps the index struct (defined in the index.go file), and store types to coordinate operations across the store and index files. This is because every time a record gets added to the store file, the index file needs to be updated with the offset and position values. For reads, the segments needs to look for the index from the index file, and search for the record at that index from the store file. The index and store files
 are saved with names that correspond to their baseOffset number; for example - with baseOffset of 3, the index and store file would be 3.index and 3.store.  This naming convention gets handy when creating new segments in the segment.go file, as the baseOffset number for the index files and store files of a particular segment can be directly parsed from the file names. Also, if you are confused, by offset number of a record, I mean the the index of a record, for example is it the first record or second or third in the store file. By baseoffSet, I mean the offset number of the first record which was written into the record file. When the index and store files of a segment reaches the max size, new index and store files are created, and the files correspond to each other; i.e one cannot be created without the other , or their information sync exactly.


## Development 
 The entire development of the project is dependent on my learning curve, and ability to grasp the concepts of distirbued services. Since, the 
 project is entirely for educational purposes, it is hard to predict a possible timeline. However, by the end of this month, the entire project can
 be expected to be completed.

## Disclaimer
As stated above, the project is being built by closely following the content and concepts outlined in the book 
Distributed Services with Go by Travis Jeffery. Hence, it's being developed purely for my own exploration and learning about distribued services.
Howvever, anyone willing to contribue is more than welcomed. Thanks!.

## Author
Hamza Yusuff - Email: hbyusuff@uwaterloo.ca

