Websites:
---
[Wikipedia](https://en.wikipedia.org/wiki/RATBV)
[Similar app](https://www.trafic-web.ro/)
[Public Transport](http://overpass-api.de/public_transport.html)
[Brasov Guide](https://www.ghid-brasov.ro/)
--

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

Boundary of Brasov (Metropolitan Area)
```
[out:json];
relation["name"="Brașov"]["type"="boundary"]["boundary"="administrative"]["place"="city"];
out body;
>;
out skel qt;
```

Import initial data
---

Download `romania-latest.osm.pbf` from https://download.geofabrik.de/europe/romania.html

`sudo apt install osmctools`

`osmconvert romania.osm.pbf -B=brasov_boundary.poly -o=brasov.osm.pbf`
