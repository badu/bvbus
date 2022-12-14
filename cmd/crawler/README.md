Definitions
---
Station - it's a bus stop, multiple Vehicle can stop in a Station by a time table. Stations have edges (links to other stations because of the routes)

Bus / Trolleybus Line - it's a Vehicle which performs a Route using a time table.

Time tables - it's the link between Station and Vehicle, useful when you need to be in the Station at a certain time

Usage
---
Run crawler if you need to get new json data.

Other interesting websites:
---
[Wikipedia](https://en.wikipedia.org/wiki/RATBV)
[Similar app](https://www.trafic-web.ro/)

[Extract of OSM Data](https://overpass-turbo.eu/)

Run with the query below, to get `geo.json`
```overpass query
[out:json][timeout:25];
(
  node({{bbox}})[network="RAT Brașov"];
);
out body;
>;
out skel qt;
```

can play with too: 
```
[out:json][timeout:25];
(
  relation["to"="Stadionul Municipal"]({{bbox}});
);
out body;
>;
out skel qt;
```

[Public Transport](http://overpass-api.de/public_transport.html)
[Brasov Guide](https://www.ghid-brasov.ro/)
--

