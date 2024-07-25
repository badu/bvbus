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

752032988
752033004
752032993
752010522
752032989
1149169825
1149169826
1149169824
752033005
752033001
 
[1] 752032988 start 7029883026 end 7029882850
[2] 752033001 start 7029882850 end 7029883029
[3] 752033005 start 7029912965 end 7029883029
[4] 1149169824 start 7029881439 end 7029912965
[5] 1149169826 start 10690954230 end 7029881439
[6] 1149169825 start 10690954230 end 7029912966
[7] 752032989 start 3357249608 end 7029912966
[8] 752010522 start 3357251368 end 3357249608
[9] 752032993 start 7029882980 end 3357251368
[10] 752033004 start 7029883026 end 7029882980
