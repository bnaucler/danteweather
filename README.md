# danteweather v0.2
What would Dante have said about the weather today?  
The service retrieves IP geolocation data, a weather report and matches with quotes from Divina Comedia. Because why not?

## Usage
Use dantewrite.go to (manually) build a database, then danteserv.go to run the web service. That's all. Global variables for editing are located in dlib/dlib.go. Once the database has been fully populated it will be included in the repo. 

## Written by
Björn Westerberg Nauclér (mail@bnaucler.se)  
Hazel Oosterhof (hazel@studiofenix.nl)

## Thanks to
Dante Alighieri  
Jessie Frazelle  
The BoltDB team  
Google

# Where can I access the service?
The server is currently running at [dante.bnaucler.se](http://dante.bnaucler.se). However, the database is yet only scarcely populated, responsive layout for tablets not yet in place, and much code optimization still needed. All in all, do not expect stability, speed nor a very accurately chosen quote.

## License
MIT (do whatever you want)
