const metroBusses = [{"i":13342503,"si":13342504,"b":"110: Cristian - Stadionul Municipal","f":"Tineretului","t":"Stadionul Municipal","n":"110","c":"#4ea74b","d":1,"s":[9182766576,9182766557,9182766561,6480402085,9186691960,3708936522,3708936521,3701493888,9182766577]},
{"i":13342504,"si":13342503,"b":"110: Stadionul Municipal - Cristian","f":"Stadionul Municipal","t":"Tineretului","n":"110","c":"#4ea74b","d":2,"s":[9182766577,3701493889,3713539337,3708962052,3713539338,9182766553,9182766555,9182766559,9182766562,9182766564,11913666846,11016370557,11913666845,9182766566,9182766568,9182766576]},
{"i":16033483,"si":16033482,"b":"120: Vulcan - Cristian - Stadionul Municipal","f":"Primaria Vulcan","t":"Stadionul Municipal","n":"120","c":"#4ea74b","d":1,"s":[11016370560,11913666845,9182766566,9182766568,9182766576,9182766557,9182766561,6480402085,9186691960,3708936522,3708936521,3701493888,9182766577]},
{"i":16033482,"si":16033483,"b":"120: Stadionul Municipal - Cristian - Vulcan","f":"Stadionul Municipal","t":"Primaria Vulcan","n":"120","c":"#4ea74b","d":2,"s":[9182766577,3701493889,3713539337,3708962052,3713539338,9182766553,9182766555,9182766559,9182766562,9182766564,11913666846,11016370557,11240465558,11016370560]},
{"i":13342988,"si":13342989,"b":"130: Rasnov - Stadionul Municipal","f":"Cap Linie Rasnov","t":"Stadionul Municipal","n":"130","c":"#4ea74b","d":1,"s":[9183241798,9183241800,9183241802,9183241804,9183241807,9183241796,9183241811,5157190621,9183241815,9182766557,9182766561,6480402085,9186691960,3708936522,3708936521,3701493888,9182766577]},
{"i":13342989,"si":13342988,"b":"130: Stadionul Municipal - Rasnov","f":"Stadionul Municipal","t":"Cap Linie Rasnov","n":"130","c":"#4ea74b","d":2,"s":[9182766577,3701493889,3713539337,3708962052,3713539338,9182766553,9182766555,9182766559,9183244618,9183244622,9183241816,4440817889,9183241812,9183241810,9183241806,9331332656,9183241803,9183241799,9183241798]},
{"i":13343061,"si":13343062,"b":"131: Romacril - Stadionul Municipal","f":"Romacril","t":"Stadionul Municipal","n":"131","c":"#4ea74b","d":1,"s":[5158419327,9183297500,9183297498,9183297494,11925977293,9183297495,9183241796,9183241811,5157190621,9183241815,9182766557,9182766561,6480402085,9186691960,3708936522,3708936521,3701493888,9182766577]},
{"i":13343062,"si":13343061,"b":"131: Stadionul Municipal - Romacril","f":"Stadionul Municipal","t":"Romacril","n":"131","c":"#4ea74b","d":2,"s":[9182766577,3701493889,3713539337,3708962052,3713539338,9182766553,9182766555,9182766559,9183244618,9183244622,9183241816,4440817889,5154009021,9183297499,5158419327]},
{"i":17802672,"si":17802671,"b":"140: CEC Zarnesti - Pepco Zarnesti - Stadionul Municipal","f":"CEC Zarnesti","t":"Stadionul Municipal","n":"140","c":"#4ea74b","d":1,"s":[11030879653,5158435528,11821207041,12125086342,5158435527,11821207044,5158419326,5158435536,4231721850,11680951479,11186302799,5158419327,9183297500,9183297498,9183241796,9183241811,5157190621,9183241815,9182766557,9182766561,6480402085,9186691960,3708936522,3708936521,3701493888,9182766577]},
{"i":17802671,"si":17802672,"b":"140: Stadionul Municipal - Penny Zarnesti - CEC Zarnesti","f":"Stadionul Municipal","t":"CEC Zarnesti","n":"140","c":"#4ea74b","d":2,"s":[9182766577,3701493889,3713539337,3708962052,3713539338,9182766553,9182766555,9182766559,9183244618,9183244622,9183241816,4440817889,5154009021,9183297499,11030867200,11186302796,11680951482,5158435537,11030879648,5158435527,11821207044,5158419326,12125086340,11821207043,5158436621,11030879653]},
{"i":13354415,"si":13354416,"b":"210: Stadionul Municipal - Ghimbav Cap Linie","f":"Stadionul Municipal","t":"Ghimbav Cap Linie","n":"210","c":"#4ea74b","d":2,"s":[9275068609,3701493889,3713539337,3708974251,9192952980,9192952984,9192952985,9192952989,9192952986,9192952991,9192952993,9192952995,9192952997]},
{"i":13354416,"si":13354415,"b":"210: Ghimbav Cap Linie - Stadionul Municipal","f":"Ghimbav Cap Linie","t":"Stadionul Municipal","n":"210","c":"#4ea74b","d":1,"s":[9192952997,9192952999,9192953001,9192952982,3708974251,3701493888,9275068609]},
{"i":13354663,"si":13354662,"b":"220: Codlea Nord - Stadionul Municipal","f":"Codlea Nord","t":"Stadionul Municipal","n":"220","c":"#4ea74b","d":1,"s":[9193088760,9193088759,9193088755,5289892947,9193088751,9192952986,3708974251,3701493888,9275068609]},
{"i":13354662,"si":13354663,"b":"220: Stadionul Municipal - Codlea Nord","f":"Stadionul Municipal","t":"Codlea Nord","n":"220","c":"#4ea74b","d":2,"s":[9275068609,3701493889,3713539337,3708974251,9192952980,9192952984,9192952985,9193088750,5289892960,5290508967,5298247922,9193088760]},
{"i":15075212,"si":15075211,"b":"310: Satu Nou cap linie - Terminal Gara","f":"Satu Nou cap linie","t":"Terminal Gara","n":"310","c":"#4ea74b","d":1,"s":[9197399856,9197399855,12043108258,10258912712,9197399850,9197399845,12043108230,9197399841,10300081327,3707831705,3707831709,3707831693,3709463826,3709463825,3709458524,3343778550,3343778546,3473944089,11801788124]},
{"i":15075211,"si":15075212,"b":"310: Terminal Gara - Satu Nou cap linie","f":"Terminal Gara","t":"Satu Nou cap linie","n":"310","c":"#4ea74b","d":2,"s":[11801788124,3343778547,3343778551,3709449858,3709449854,9172134414,3709449855,3707786426,9171304342,3707786436,3707786432,10300081325,9197399840,12043108232,9197399844,9197399848,10258912709,12043108260,9197399852,9197399856]},
{"i":17580531,"si":17580530,"b":"320: Rotbav - Terminal Gara","f":"Rotbav","t":"Terminal Gara","n":"320","c":"#4ea74b","d":1,"s":[9197486573,9197486569,417677566,9197486567,9197486560,10064979821,10300081327,3707831705,3707831709,3707831693,3709463826,3709463825,3709458524,3343778550,3343778546,3473944089,11801788122]},
{"i":17580530,"si":17580531,"b":"320: Terminal Gara - Rotbav","f":"Terminal Gara","t":"Rotbav","n":"320","c":"#4ea74b","d":2,"s":[11801788122,3343778547,3343778551,3709449858,3709449854,9172134414,3709449855,3707786426,9171304342,3707786436,3707786432,10300081325,10064979823,9197486562,417677566,9197486565,9197486567,9197486571,9197486573]},
{"i":13365595,"si":13365596,"b":"410: Subcetate - Rulmentul","f":"Subcetate","t":"Rulmentul","n":"410","c":"#4ea74b","d":1,"s":[9198172121,9198172120,9197991345,9275045758]},
{"i":13365596,"si":13365595,"b":"410: Rulmentul - Subcetate","f":"Rulmentul","t":"Subcetate","n":"410","c":"#4ea74b","d":2,"s":[9275045758,9197991344,9198172118,9198172121]},
{"i":13365770,"si":13365769,"b":"411: Rulmentul - Residence (capat)","f":"Rulmentul","t":"Residence (capat)","n":"411","c":"#4ea74b","d":2,"s":[9275045758,9198279353,9198279357,9198279361,9331334987,9331334988,9331334985]},
{"i":13365769,"si":13365770,"b":"411: Residence (capat) - Rulmentul","f":"Residence (capat)","t":"Rulmentul","n":"411","c":"#4ea74b","d":1,"s":[9331334985,9331334990,12085098229,12085098231,12085098233,9199390733,9199390729,9198279358,9198279354,9275045758]},
{"i":13367388,"si":13367387,"b":"412: Rulmentul - Spital","f":"Rulmentul","t":"Morii Spital","n":"412","c":"#23b14d","d":2,"s":[9275045758,9198279353,9198279357,9199390727,9199390731,9199390735,9199390738,12085098240,12085098242,12085098245]},
{"i":13367387,"si":13367388,"b":"412: Spital Sanpetru - Rulmentul","f":"Spital Sanpetru","t":"Rulmentul","n":"412","c":"#23b14d","d":1,"s":[12085098245,12085098244,12085098238,9199390742,9199390740,9199390737,9199390733,9199390729,9198279358,9198279354,9275045758]},
{"i":13370055,"si":13369963,"b":"420: Rulmentul - Bod Colonie","f":"Rulmentul","t":"Bod Colonie","n":"420","c":"#4ea74b","d":2,"s":[9275045757,9198279353,9198279357,9199390727,9199390731,9199390735,9199390738,9200790448,9200790452,9200790455,9200790460,9200790463,8123995911,9200790467,9200790470,11428570047,9200790474,9200790478,677334149]},
{"i":13369963,"si":13370055,"b":"420: Bod Colonie - Rulmentul","f":"Bod Colonie","t":"Rulmentul","n":"420","c":"#4ea74b","d":1,"s":[677334149,9200790479,9200790475,11428570044,9200790471,9200790468,8123995911,9200790464,9200790459,9200790456,9200790454,9200790449,9199390740,9199390737,9199390733,9199390729,9198279358,9198279354,9275045757]},
{"i":17580529,"si":17580528,"b":"511: Cap Linie Podu Oltului - Terminal Gara","f":"Cap Linie Podu Oltului","t":"Terminal Gara","n":"511","c":"#4ea74b","d":1,"s":[9212174593,9212174591,9212174587,9212174583,9212174579,9212174575,11159652426,11133798595,9212097842,2088441880,2088438980,9212097837,9188081130,3407053698,3652393629,3708943734,11801788124]},
{"i":17580528,"si":17580529,"b":"511: Terminal Gara - Cap Linie Podu Oltului","f":"Terminal Gara","t":"Cap Linie Podu Oltului","n":"511","c":"#4ea74b","d":2,"s":[11801788124,2910431428,3652436632,3652436626,3407053697,9212097836,2088438287,2088442014,9212174573,9212174577,9212174581,9212174585,9212174589,9212174593]},
{"i":17580527,"si":17580526,"b":"520: Lunca Calnicului (Vizaviu) - Terminal Gara","f":"Lunca Calnicului (Vizaviu)","t":"Terminal Gara","n":"520","c":"#4ea74b","d":1,"s":[9214351456,9214351454,9214351451,9214351449,9214351445,9214351442,9214351440,9228458196,9214351433,5648409764,9214351426,7158943526,9214351418,7159000047,2088363037,9274917904,9274917903,3407053698,3652393629,3708943734,11801788126]},
{"i":17580526,"si":17580527,"b":"520: Terminal Gara - Lunca Calnicului (Vizaviu)","f":"Terminal Gara","t":"Lunca Calnicului (Vizaviu)","n":"520","c":"#4ea74b","d":2,"s":[11801788126,3473944089,2910431428,3652436632,3652436626,3407053697,2088438290,7159000045,9214351420,9214351422,9214351424,9214351428,9214351431,9214351435,9214351438,9214351441,9214351444,9214351447,9214351450,9214351453,9214351456]},
{"i":17580524,"si":17580525,"b":"540: Vama de Sus - Terminal Gara","f":"Vama de Sus","t":"Terminal Gara","n":"540","c":"#4ea74b","d":1,"s":[9214436153,9214436150,9214436147,9214436144,9214436141,9214436138,9214436135,9214436132,9214436129,9214436126,9956538861,9956538857,9956538811,9433605252,9956538787,9433605250,9433605245,9214436124,9214436121,9214436119,9214435716,7158943526,7159000047,2088363037,9188081130,3407053698,3652393629,3708943734,11801788127]},
{"i":17580525,"si":17580524,"b":"540: Terminal Gara - Vama de Sus","f":"Terminal Gara","t":"Vama de Sus","n":"540","c":"#4ea74b","d":2,"s":[11801788127,2910431428,3652436632,3652436626,3407053697,2088438290,7159000045,9214351422,9214435715,9214436118,9214436122,9433605247,9956425089,9433605249,9956538786,9433605254,9956538809,9956538823,9956538859,9214436125,9214436130,9214436134,9214436136,9214436140,9214436143,9214436145,9214436149,9214436152,9214436153]},
{"i":13393598,"si":13393597,"b":"610: Cap Linie Purcareni - Cap Linie Tarlungeni - Roman","f":"Cap Linie Purcareni","t":"Roman","n":"610","c":"#4ea74b","d":1,"s":[9215921887,9215921886,9215921883,9215921877,9215921875,9215921871,9215921868,9215921867,9215921864,9215921860,9215921856,9164143090,3709393825,2657652503,2657624300,285721074,3709393832,2537999895,9274823442]},
{"i":13393597,"si":13393598,"b":"610: Roman - Cap Linie Tarlungeni - Cap Linie Purcareni","f":"Roman","t":"Cap Linie Purcareni","n":"610","c":"#4ea74b","d":2,"s":[9274823442,2537998221,3709393831,2657617744,2657624828,2657653187,2657655619,9168992171,9215921854,9215921858,9215921862,9215921866,9215921869,9215921873,9215921875,9215921876,9215921879,4726062976,9215921881,9215921884,9215921887]},
{"i":13393664,"si":13393665,"b":"611: Roman - Cap Linie Tarlungeni","f":"Roman","t":"Cap Linie Tarlungeni","n":"611","c":"#4ea74b","d":2,"s":[9274823442,2537998221,3709393831,2657617744,2657624828,2657653187,2657655619,9168992171,9215921854,9215921889,9215921891,9215921895,9215921893,9215921858,9215921862,9215921866,9215921869,9215921873,9215921875]},
{"i":13393665,"si":13393664,"b":"611: Cap Linie Tarlungeni - Roman","f":"Cap Linie Tarlungeni","t":"Roman","n":"611","c":"#4ea74b","d":1,"s":[9215921875,9215921871,9215921868,9215921867,9215921864,9215921889,9215921891,9215921895,9215921893,9215921860,9215921856,9164143090,3709393825,2657652503,2657624300,285721074,3709393832,2537999895,9274823442]},
{"i":13393856,"si":13393857,"b":"612: Roman - Cap Linie Purcareni","f":"Roman","t":"Cap Linie Purcareni","n":"612","c":"#4ea74b","d":2,"s":[9274823442,2537998221,3709393831,2657617744,2657624828,2657653187,2657655619,9168992171,9215921854,9215921889,9215921891,9215921895,9215921893,9215921858,9215921862,9215921866,9215921869,9215921873,9215921875,9215921871,9215921868,9215921867,9215921876,9215921879,4726062976,9215921881,9215921884,9215921887]},
{"i":13393857,"si":13393856,"b":"612: Cap Linie Purcareni - Roman","f":"Cap Linie Purcareni","t":"Roman","n":"612","c":"#4ea74b","d":1,"s":[9215921887,9215921886,9215921883,9215921877,9215921864,9215921860,9215921856,9164143090,3709393825,2657652503,2657624300,285721074,3709393832,2537999895,9274823442]},
{"i":17132991,"si":17132992,"b":"620: Cap Linie Budila - Gemenii","f":"Cap Linie Budila","t":"Gemenii","n":"620","c":"#4ea74b","d":1,"s":[9215921902,9215921901,9215921898,9215921864,9215921860,9215921856,9164143090,3709393825,2657652503,2657624300,285721074,3709393832]},
{"i":17132992,"si":17132991,"b":"620: Gemenii - Cap Linie Budila","f":"Gemenii","t":"Cap Linie Budila","n":"620","c":"#4ea74b","d":2,"s":[3709393831,2657617744,2657624828,2657653187,2657655619,9168992171,9215921854,9215921858,9215921862,9215921897,9215921900,9215921902]},
{"i":13545862,"si":13545863,"b":"710: Garaje Sacele - Roman","f":"Garaje Sacele","t":"Roman","n":"710","c":"#4ea74b","d":1,"s":[9334490424,301798666,9334490420,9334490419,9334399616,303078870,9334399610,9334399608,9334399604,9334399598,3709600097,611633525,611633519,9187345296]},
{"i":13545863,"si":13545862,"b":"710: Roman - Garaje Sacele","f":"Roman","t":"Garaje Sacele","n":"710","c":"#4ea74b","d":2,"s":[9187345296,1864408175,3701600885,3709600098,9334399599,9334399602,9334399606,9334399612,9334399613,9334399614,9334490417,9334490421,301798665,10610528741]},
{"i":16624359,"si":16624358,"b":"711: Roman - Garcini","f":"Roman","t":"Garcini","n":"711","c":"#4ea74b","d":2,"s":[9187345296,1864408175,3701600885,3709600098,9334399599,9334399602,9334399606,9334399612,9334399613,9334399614,9334490417,9334490421,301798665,10610528741,9334490426]},
{"i":16624358,"si":16624359,"b":"711: Garcini - Roman","f":"Garcini","t":"Roman","n":"711","c":"#4ea74b","d":1,"s":[9334490426,9334490424,301798666,9334490420,9334490419,9334399616,303078870,9334399610,9334399608,9334399604,9334399598,3709600097,611633525,611633519,9187345296]},
{"i":13396980,"si":13396981,"b":"810: Roman - Cap Linie Predeal","f":"Roman","t":"Cap Linie Predeal","n":"810","c":"#4ea74b","d":2,"s":[9274823443,1864408175,3701600885,3709600098,3709600108,3709600102,3709600104,3709600106,3709600100,9216522092,9216522096,9216522099,9216522101,9216522104,9216522108,9216522110,9216522116,9216527118]},
{"i":13396981,"si":13396980,"b":"810: Cap Linie Predeal - Roman","f":"Cap Linie Predeal","t":"Roman","n":"810","c":"#4ea74b","d":1,"s":[9216527118,9216522114,9216522112,9216522107,9216522105,9216522104,9216522102,9216522098,9216522097,9216522094,9216522091,3709600099,3709600105,3709600103,3709600101,3709600107,3709600097,611633525,611633519,9274823443]},
]
export default metroBusses;
