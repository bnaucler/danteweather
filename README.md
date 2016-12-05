# danteweather v0.2
What would Dante have said about the weather today?  
The service retrieves IP geolocation data, a weather report and matches with quotes from Divina Comedia. Because why not?

## Usage
`go run danteserv` will start the server at port 8959. That's all. Global variables for editing are located in dlib/dlib.go. Once the database has been fully populated it will be included in the repo.

The helper applications 'dantewrite' and 'dantedump' are used to build and view the database, respectively. They are not very user friendly, but does the job, and can be used if anyone wants to build a similar service, quoting someone or something else.

## Written by
Björn Westerberg Nauclér (mail@bnaucler.se)  
Hazel Oosterhof (hazel@studiofenix.nl)

## Thanks to
Dante Alighieri  
Jessie Frazelle  
The BoltDB team  
Google

## Where can I access the service?
The server is currently running at [dante.bnaucler.se](http://dante.bnaucler.se). However, the database is yet only scarcely populated, responsive layout for tablets not yet in place, and much code optimization still needed. All in all, do not expect stability, speed nor a very accurately chosen quote.

## License
MIT (do whatever you want)
