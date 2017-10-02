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
The project is currenly put on hold and the temporary instance shut down. It is unclear at this time if it will be completed.

## License
MIT (do whatever you want)
