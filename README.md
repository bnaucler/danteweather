# danteweather v0.2
What would Dante have said about the weather today?  
The service retrieves IP geolocation data, a weather report and matches with quotes from Divina Comedia. Because why not?

## Usage
Use dantewrite.go to (manually) build a database, then danteserv.go to run the web service. That's all. Global variables for editing are located in dlib/dlib.go. Once the database has been fully populated it will be included in the repo. (There's no point at the moment as it's filled with changing test data.)

## Written by
Björn Westerberg Nauclér (mail@bnaucler.se)  
Hazel Oosterhof (hazel@studiofenix.nl)

## Thanks to
Dante Alighieri  
Jessie Frazelle  
The BoltDB team  
Google

# Where can I access the service?
The application is in an early beta state; unstable and not really useful. Hence there are only temporary instances set up for specific tests. This will change once the database has been populated.

## License
MIT (do whatever you want)
