const busses = [{"i":5369802,"b":"1: Livada Postei - Triaj","f":"Livada Postei","t":"Triaj","n":"1","c":"#ffe900","d":1,"s":[2375041371,2655859050,262148563,8227802075,3701356197,465241947,8313496897,3652436632,3652436626,3407053697,9274917899]},
{"i":5369803,"b":"1: Triaj - Livada Postei","f":"Triaj","t":"Livada Postei","n":"1","c":"#ffe900","d":2,"s":[9274917899,3407053698,3652393629,300089533,300089535,465241917,267042578,3652393631,2537929578,2552635273,2375041371]},
{"i":5369951,"b":"2: Livada Postei - Rulmentul","f":"Livada Postei","t":"Rulmentul","n":"2","c":"#00b64f","d":1,"s":[9183614613,2655859050,2679017945,1923401916,3652479574,3368804794,3652479577,3652479573,3713477178,3652479575,2657605414]},
{"i":5369952,"b":"2: Rulmentul - Livada Postei","f":"Rulmentul","t":"Livada Postei","n":"2","c":"#00b64f","d":2,"s":[2657605414,3652479576,3713477179,3652479572,2657703726,3473944193,1923401922,2537929578,2552635273,9183614613]},
{"i":12995687,"b":"2B: Rulmentul - Ioan Popasu - Livada Postei","f":"Rulmentul","t":"Livada Postei","n":"2B","c":"#00b64f","d":2,"s":[9275045759,8934219730,8934219727,8934219725,567945864,3713477179,3652479572,2657703726,3473944193,1923401922,2537929578,2552635273,2375041371]},
{"i":12995686,"b":"2B: Livada Postei - Ioan Popasu - Rulmentul","f":"Livada Postei","t":"Rulmentul","n":"2B","c":"#00b64f","d":1,"s":[2375041371,2655859050,2679017945,1923401916,3652479574,3368804794,3652479577,3652479573,3713477178,8934219721,8934219723,8934219728,8934219729,9275045759]},
{"i":5417774,"b":"3: Stadionul Tineretului - Valea Cetatii","f":"Stadionul Tineretului","t":"Valea Cetatii","n":"3","c":"#a6aca2","d":1,"s":[3474036169,3701534855,3343778544,3343778549,3343778552,1923401922,262148563,8227802075,3640849654,2910858095,272095074,272095075,272095077,272095079,314157693,353099740,2657492053]},
{"i":5417775,"b":"3: Valea Cetatii - Stadionul Tineretului","f":"Valea Cetatii","t":"Stadionul Tineretului","n":"3","c":"#a6aca2","d":2,"s":[3713443720,353100201,264561536,290004759,273437289,307019142,3701356196,3701356197,267042578,3343781090,3343778553,3343778551,3347300843,3474036169]},
{"i":14280746,"b":"4: Tocile - Terminal Gara","f":"Tocile","t":"Terminal Gara","n":"4","c":"#7da1c5","d":2,"s":[3713320058,3713320059,3713320056,2375041369,2655859050,262148563,8227802075,2683094124,2683092725,10198827064]},
{"i":14280747,"b":"4: Terminal Gara - Tocile","f":"Terminal Gara","t":"Tocile","n":"4","c":"#7da1c5","d":1,"s":[10198827064,2683090885,2683091443,267042578,3652393631,2537929578,2552635273,2375041368,3713320054,3713320055,2854669944,3713320058]},
{"i":5397475,"b":"5: Stadionul Municipal - Roman","f":"Stadionul Municipal","t":"Roman","n":"5","c":"#ec2738","d":1,"s":[9275068611,3701356185,3701356188,3701356191,3701356186,3701356184,2655859050,262148563,8227802075,3640849654,310018170,3701356190,2657677035,3701356193,3713302845,2680755471,9274823441]},
{"i":17828247,"b":"5: Roman - De Mijloc - Stadionul Municipal","f":"Roman","t":"Stadionul Municipal","n":"5","c":"#ec2738","d":2,"s":[9274823441,3701356195,611633526,611633527,3701356196,3701356197,267042578,3652393631,2537929578,2552635273,9245702218,12010293521,12010293524,12010293526,12010293528,3654024249,9275068611]},
{"i":5372252,"b":"5M: Stadionul Municipal - Magurele","f":"Stadionul Municipal","t":"Magurele","n":"5M","c":"#e77721","d":2,"s":[9275068609,3701493889,3713539337,3713539341,3713539338,3714980425,3713539336,3714980427,3713539342]},
{"i":5372251,"b":"5M: Magurele - Stadionul Municipal","f":"Magurele","t":"Stadionul Municipal","n":"5M","c":"#e77721","d":1,"s":[3713539342,3713539340,3713539339,3708936522,3708936521,3701493888,9275068609]},
{"i":5372281,"b":"6: Saturn - Livada Postei","f":"Saturn","t":"Livada Postei","n":"6","c":"#00b298","d":1,"s":[3701409066,3701409064,310840809,2537998221,2537995437,464246419,267042578,3652393631,2537929578,2552635273,2375041371]},
{"i":5372280,"b":"6: Livada Postei - Saturn","f":"Livada Postei","t":"Saturn","n":"6","c":"#00b298","d":2,"s":[2375041371,2655859050,262148563,8227802075,264864013,3653874651,2537994014,2537999895,2538006527,2538009452,3701409066]},
{"i":5417865,"b":"7: Rulmentul - Roman","f":"Rulmentul","t":"Roman","n":"7","c":"#ff6900","d":2,"s":[2657605414,3652479576,3713477179,3652479572,2657703726,3473944193,1923401922,262148563,8227802075,3640849654,2910858095,272095074,2680733271,3713302845,2680755471,9274823444]},
{"i":5417864,"b":"7: Roman - Rulmentul","f":"Roman","t":"Rulmentul","n":"7","c":"#ff6900","d":1,"s":[9274823444,2680751671,611633518,611633514,307019142,3701356196,3701356197,267042578,1923401916,3652479574,3368804794,3652479577,3652479573,3713477178,3652479575,2657605414]},
{"i":5417963,"b":"8: Rulmentul - Saturn","f":"Rulmentul","t":"Saturn","n":"8","c":"#e5bedd","d":1,"s":[2657605414,3652479576,3713477179,3652479572,2657703726,3473944089,2910431428,464276377,299937126,2537994014,2537999895,2538006527,2538009452,9274932348]},
{"i":5417964,"b":"8: Saturn - Rulmentul","f":"Saturn","t":"Rulmentul","n":"8","c":"#e5bedd","d":2,"s":[9274932348,3701409064,310840809,2537998221,2537995437,2537991836,464267110,3708943734,1635108121,3368804794,3652479577,3652479573,3713477178,3652479575,2657605414]},
{"i":5372431,"b":"9: Stadionul Municipal - Rulmentul","f":"Stadionul Municipal","t":"Rulmentul","n":"9","c":"#eee04b","d":2,"s":[9275068610,3654024250,3654024251,3654024253,2657710288,3654024255,3652479577,3652479573,3713477178,8934219721,8934219723,8934219728,8934219729,9275045760]},
{"i":5372432,"b":"9: Rulmentul - Stadionul Municipal","f":"Rulmentul","t":"Stadionul Municipal","n":"9","c":"#eee04b","d":1,"s":[9275045760,8934219730,8934219727,8934219725,567945864,3713477179,3652479572,2657704969,2657706440,2657713531,3654024254,3654024252,3654024249,9275068610]},
{"i":5417974,"b":"10: Triaj - Valea Cetatii","f":"Triaj","t":"Valea Cetatii","n":"10","c":"#9e292f","d":1,"s":[9274917899,3407053698,3652393629,300089533,300089535,465241917,267042578,8227802075,3640849654,2910858095,272095074,272095075,272095077,272095079,314157693,353099740,2657492053]},
{"i":5417975,"b":"10: Valea Cetatii - Triaj","f":"Valea Cetatii","t":"Triaj","n":"10","c":"#9e292f","d":2,"s":[3713443720,353100201,264561536,290004759,273437289,307019142,3701356196,3701356197,465241947,8313496897,3652436632,3652436626,3407053697,9274917899]},
{"i":17828246,"b":"14: Livada Postei - De Mijloc - Fabrica de Var","f":"Livada Postei","t":"Fabrica de Var","n":"14","c":"#6dc24b","d":2,"s":[2375041369,9245702218,12010293521,12010293524,12010293526,12010293528,3701356185,3709521411,3709521417,3709521421,3709521423,3709521415,5372688332,3709521410,3709521412]},
{"i":5399072,"b":"14: Fabrica de Var - Livada Postei","f":"Fabrica de Var","t":"Livada Postei","n":"14","c":"#6dc24b","d":1,"s":[3709521412,3709521409,9245675516,3709521414,3709521422,3709521420,3709521416,3701356188,3701356191,3701356186,3701356184,2375041369]},
{"i":5409348,"b":"15: Triaj - Avantgarden","f":"Triaj","t":"Avantgarden","n":"15","c":"#06048c","d":2,"s":[9274917903,3407053698,3709554454,3709554456,8989955187,3709554464,3652479572,2657704969,2657706440,2657713531,3654024254,3654024252,3708991157,3708991160,3708991159,3708991158,9170142236,5562037662]},
{"i":5409347,"b":"15: Avantgarden - Triaj","f":"Avantgarden","t":"Triaj","n":"15","c":"#06048c","d":1,"s":[5562037662,5562037663,3709002258,3709002260,3709045329,3709002261,3709291642,3654024251,3654024253,2657710288,3654024255,3652479577,3652479573,3709554462,3709554455,3709554453,3407053697,9274917903]},
{"i":5386103,"b":"16: Livada Postei - Stadionul Municipal","f":"Livada Postei","t":"Stadionul Municipal","n":"16","c":"#fa6544","d":2,"s":[2375041372,2655859050,2679017945,3343781090,3343778553,3343778551,3347300843,5218033007,3654024249,3701493889,3701493888,9182766577]},
{"i":5386102,"b":"16: Stadionul Municipal - Livada Postei","f":"Stadionul Municipal","t":"Livada Postei","n":"16","c":"#fa6544","d":1,"s":[9182766577,3654024250,3474036169,3701534855,3343778544,3343778549,3343778552,1923401922,2537929578,2552635273,2375041372]},
{"i":5386136,"b":"17: Noua - Livada Postei","f":"Noua","t":"Livada Postei","n":"17","c":"#6ad1e2","d":2,"s":[2657574024,611633517,611633516,611633522,611633524,611633525,611633519,3701356195,611633526,611633527,3701356196,3701356197,267042578,3652393631,2537929578,2552635273,2375041369]},
{"i":5386137,"b":"17: Livada Postei - Noua","f":"Livada Postei","t":"Noua","n":"17","c":"#6ad1e2","d":1,"s":[2375041369,2655859050,262148563,8227802075,3640849654,310018170,3701356190,2657675435,3701600887,1864408175,3701600885,611633523,611633521,611633515,2657569557,2657570995,2657574024]},
{"i":5409920,"b":"17B: Timisul de Jos - Terminal Gara","f":"Benzinaria Petrom","t":"Terminal Gara","n":"17B","c":"#ffc72c","d":1,"s":[3709600099,3709600105,3709600103,3709600101,3709600107,11893411924,3709600097,611633525,611633519,9274932345,3701409064,310840809,2537998221,2537995437,2537991836,464267110,3708943734,11801788121]},
{"i":5409919,"b":"17B: Terminal Gara -Timisul de Jos","f":"Terminal Gara","t":"Benzinaria Petrom","n":"17B","c":"#ffc72c","d":2,"s":[11801788121,2910431428,464276377,299937126,2537994014,2537999895,2538006527,2538009452,3708904920,1864408175,3701600885,3709600098,11893411924,3709600108,3709600102,3709600104,3709600106,3709600099]},
{"i":13338246,"b":"18: Bariera Bartolomeu - IAR Ghimbav","f":"Bariera Bartolomeu","t":"IAR Ghimbav","n":"18","c":"#e58799","d":2,"s":[3707775824,3707786431,3707786425,3707786437,3707786433,3707786426,9171304342,3707786436,3707786432,3707786429,3707786430,3707786428,3707786427,3713233437,3707786435,3707795551,3707795550,3709045334,3707795549,9171277177]},
{"i":13338245,"b":"18: IAR Ghimbav - Bariera Bartolomeu","f":"IAR Ghimbav","t":"Bariera Bartolomeu","n":"18","c":"#e58799","d":1,"s":[9171277177,3708889555,8926599968,3708889565,3708889558,3708889561,9170035655,3708889572,3707831699,3707831700,3707831703,3707831701,3707831705,3707831709,3707831693,3708889569,3708889575,3708889556,3708889567,3707775824]},
{"i":5387307,"b":"20: Poiana Brasov - Livada Postei","f":"Poiana Brasov","t":"Livada Postei","n":"20","c":"#96d700","d":2,"s":[3708889570,3708897319,3708897315,3708897316,300099701,3708897317,3652436629]},
{"i":5387306,"b":"20: Livada Postei - Poiana Brasov","f":"Livada Postei","t":"Poiana Brasov","n":"20","c":"#96d700","d":1,"s":[3652436629,3708889560,300099706,3708889559,3708889557,3708889571,3708889570]},
{"i":14428424,"b":"20B: Livada Postei - Belvedere","f":"Livada Postei","t":"Belvedere","n":"20B","c":"#ceef6c","d":2,"s":[3652436629,2854669944,2655859050,2552635273,9245702218,3701356185,9933396271]},
{"i":14428467,"b":"20B: Belvedere - Livada Postei","f":"Belvedere","t":"Livada Postei","n":"20B","c":"#ceef6c","d":1,"s":[9933396271,3652436629]},
{"i":5386246,"b":"21: Noua - Triaj","f":"Noua","t":"Triaj","n":"21","c":"#fc4c01","d":1,"s":[2657574024,611633517,611633516,611633522,611633524,611633525,611633519,9274932347,3701409064,310840809,2537998221,2537995437,2537991836,464267110,3652436632,3652436626,3407053697,9274917901]},
{"i":5386247,"b":"21: Triaj - Noua","f":"Triaj","t":"Noua","n":"21","c":"#fc4c01","d":2,"s":[9274917901,3407053698,3652393629,464276377,299937126,2537994014,2537999895,2538006527,2538009452,3708904920,1864408175,3701600885,611633523,611633521,611633515,2657569557,2657570995,2657574024]},
{"i":5387343,"b":"22: Saturn - Stadionul Tineretului","f":"Saturn","t":"Stadionul Tineretului","n":"22","c":"#bb29ba","d":2,"s":[9274932349,3701409064,310840809,2537998221,2537995437,464246419,3701356197,267042578,1923401922,2537929578,3708912144,3708912143,3708912142,3708912141,3708932945]},
{"i":5387344,"b":"22: Stadionul Tineretului - Saturn","f":"Stadionul Tineretului","t":"Saturn","n":"22","c":"#bb29ba","d":1,"s":[3708932945,3708932940,3708932942,3708932943,3708932944,3708932941,262148563,8227802075,264864013,3653874651,2537994014,2537999895,2538006527,2538009452,9274932349]},
{"i":5388542,"b":"23: Depozite ILF - Saturn","f":"Depozite ILF","t":"Saturn","n":"23","c":"#005ebe","d":1,"s":[3708936522,3708936521,3701493888,9275068612,3654024250,3474036169,3701534855,3343778544,3343778550,3343778546,3473944089,2910431428,464276377,299937126,2537994014,2537999895,2538006527,2538009452,9274932346]},
{"i":5388543,"b":"23: Saturn - Depozite ILF","f":"Saturn","t":"Depozite ILF","n":"23","c":"#005ebe","d":2,"s":[9274932346,3701409064,310840809,2537998221,2537995437,2537991836,464267110,3708943734,1635108121,3343778547,3343778551,3347300843,5218033007,3654024249,3701493889,3713539337,3708943733,3708936522]},
{"i":5388613,"b":"23B: Triaj - Stadionul Municipal","f":"Triaj","t":"Stadionul Municipal","n":"23B","c":"#db3eb1","d":2,"s":[9274917902,3407053698,3652393629,3708943734,1635108121,3473944193,3343781090,3343778553,3343778551,3347300843,5218033007,3654024249,9275068611]},
{"i":5388612,"b":"23B: Stadionul Municipal - Triaj","f":"Stadionul Municipal","t":"Triaj","n":"23B","c":"#db3eb1","d":1,"s":[9275068611,3654024250,3474036169,3701534855,3343778544,3343778549,3343778552,1923401916,3652479574,3473944089,2910431428,3652436632,3652436626,3407053697,9274917902]},
{"i":17828245,"b":"24: Livada Postei - De Mijloc - Baciului","f":"Livada Postei","t":"Baciului","n":"24","c":"#7b2854","d":2,"s":[2375041372,9245702218,12010293521,12010293524,12010293526,12010293528,3654024249,3701493889,3713539337,3708962052,3708962054,3708962050,3708962051,9710744128,8926599968,3708889565,3708889558,3708889561,9170035655,3707786434,3713233439,3713232814,3713232822,3713232816,3713232820,3713232802,3713232818,3713232805,9267236358]},
{"i":13337513,"b":"24: Baciului - Stupinii Noi - Livada Postei","f":"Baciului","t":"Livada Postei","n":"24","c":"#7b2854","d":1,"s":[9267236358,9267236357,3713232819,3713232803,9179480454,3713232821,3713232817,3713232823,3713232815,3831052316,3713233437,3707786435,3707795551,3707795550,3709045334,9710744131,9171391885,3708974252,3708974250,3708974254,3708974253,3708974251,3701493888,9275068612,3701356185,3701356188,3701356191,3701356186,3701356184,2375041372]},
{"i":5389640,"b":"25: Roman - Avantgarden","f":"Roman","t":"Avantgarden","n":"25","c":"#a4e1b5","d":1,"s":[9187345290,611633519,3701409064,310840809,2537998221,2537995437,2537991836,464267110,3708943734,1635108121,3343778547,3343778551,3347300843,5218033007,3708991157,3708991160,3708991159,3708991158,9170142236,5562037662]},
{"i":5389639,"b":"25: Avantgarden - Roman","f":"Avantgarden","t":"Roman","n":"25","c":"#a4e1b5","d":2,"s":[5562037662,5562037663,3709002258,3709002260,3709045329,3709002261,3654024250,3474036169,3701534855,3343778544,3343778550,3343778546,3473944089,2910431428,464276377,299937126,2537994014,2537999895,2538006527,2538009452,3708904920,9187345290]},
{"i":13338406,"b":"28: IAR Ghimbav - Livada Postei","f":"IAR Ghimbav","t":"Livada Postei","n":"28","c":"#fbda66","d":1,"s":[9171277177,3708889555,9710744131,3708889565,3709053148,3709053152,3709053149,3709053150,3709053153,3708991159,3708991158,9170142236,5562037662,3708974253,3708974251,3701493888,9275068610,3701356185,3701356188,3701356191,3701356186,3701356184,2375041372]},
{"i":17828242,"b":"28: Livada Postei - De Mijloc - IAR Ghimbav","f":"Livada Postei","t":"IAR Ghimbav","n":"28","c":"#fbda66","d":2,"s":[2375041372,9245702218,12010293521,12010293524,12010293526,12010293528,3654024249,3701493889,3713539337,3708962052,3708962054,5562037663,3709002258,3709002260,3709045329,3709045337,3709045332,3709045331,3709045336,3709045330,3709045334,3707795549,9171277177]},
{"i":16198891,"b":"29: Bartolomeu Nord - Terminal Gara","f":"Bartolomeu Nord","t":"Terminal Gara","n":"29","c":"#ef3341","d":1,"s":[3709053149,3709053150,3709053153,3709002261,3709291642,3654024251,3654024253,2657710288,3654024255,2657703726,3473944089,11801788123]},
{"i":16198892,"b":"29: Terminal Gara - Bartolomeu Nord","f":"Terminal Gara","t":"Bartolomeu Nord","n":"29","c":"#ef3341","d":2,"s":[11801788123,3368804794,2657704969,2657706440,2657713531,3654024254,3654024252,3708991157,3708991160,3709045337,3709045332,3709045331]},
{"i":5390264,"b":"31: Livada Postei - Valea Cetatii","f":"Livada Postei","t":"Valea Cetatii","n":"31","c":"#ff5001","d":1,"s":[9183614613,2655859050,262148563,8227802075,3640849654,2910858095,272095074,272095075,272095077,272095079,314157693,353099740,2657492053]},
{"i":5390265,"b":"31: Valea Cetatii - Livada Postei","f":"Valea Cetatii","t":"Livada Postei","n":"31","c":"#ff5001","d":2,"s":[3713443720,353100201,264561536,290004759,273437289,307019142,3701356196,3701356197,267042578,3652393631,2537929578,2552635273,9183614613]},
{"i":16218665,"b":"32: Valea Cetatii - 13 Decembrie","f":"Valea Cetatii","t":"13 Decembrie","n":"32","c":"#023c9f","d":2,"s":[3713443720,353100201,264561536,290004759,273437289,307019142,3701356196,3701356197,2683094124,2683092725,1635108121,3368804794,3652479577,10775488000,10774754860,3709554454,3709554456,3709554464]},
{"i":16218666,"b":"32: 13 Decembrie - Valea Cetatii","f":"13 Decembrie","t":"Valea Cetatii","n":"32","c":"#023c9f","d":1,"s":[3709554464,3652479572,2657703726,3473944089,11801788119,2683090885,2683091443,267042578,8227802075,3640849654,2910858095,272095074,272095075,272095077,272095079,314157693,353099740,2657492053]},
{"i":5418079,"b":"33: Valea Cetatii - Roman","f":"Valea Cetatii","t":"Roman","n":"33","c":"#c8c8c6","d":2,"s":[3713443720,353100201,264561536,3713514691,3713514685,2680733271,3713302845,2680755471,9274823444]},
{"i":5418078,"b":"33: Roman - Valea Cetatii","f":"Roman","t":"Valea Cetatii","n":"33","c":"#c8c8c6","d":1,"s":[9274823444,2680751671,611633518,611633514,3713514692,3713514686,272095079,314157693,353099740,2657492053]},
{"i":5390289,"b":"34: Livada Postei - Timis Triaj","f":"Livada Postei","t":"Timis Triaj","n":"34","c":"#98999b","d":2,"s":[9183614613,2655859050,262148563,8227802075,264864013,3653874651,3709393831,2657617744,2657624828,2657653187,2657655619,3709393830,3709393829,3709393836,3709393827,3709338123,3709393833,3709393837]},
{"i":5390288,"b":"34: Timis Triaj - Livada Postei","f":"Timis Triaj","t":"Livada Postei","n":"34","c":"#98999b","d":1,"s":[3709393837,3709393834,3709338124,3709393826,3709393835,3709393828,3709393825,2657652503,2657624300,285721074,3709393832,464246419,3701356197,267042578,3652393631,2537929578,2552635273,9183614613]},
{"i":5390300,"b":"34B: Izvor - Livada Postei","f":"Izvor Cap Linie","t":"Livada Postei","n":"34B","c":"#eebae1","d":1,"s":[9164143097,11817868371,9164260788,9164260789,9164143090,3709393825,2657652503,2657624300,285721074,3709393832,464246419,3701356197,267042578,3652393631,2537929578,2552635273,9183614613]},
{"i":5390299,"b":"34B: Livada Postei - Izvor","f":"Livada Postei","t":"Izvor Cap Linie","n":"34B","c":"#eebae1","d":2,"s":[9183614613,2655859050,262148563,8227802075,264864013,3653874651,3709393831,2657617744,2657624828,2657653187,2657655619,9164143091,9164143092,9164143093,9164143097]},
{"i":5390328,"b":"35: Terminal Gara - Noua","f":"Terminal Gara","t":"Noua","n":"35","c":"#b095a6","d":1,"s":[11801788121,2683090885,2683091443,267042578,8227802075,3640849654,310018170,3701356190,2657675435,3701600887,1864408175,3701600885,611633523,611633521,611633515,2657569557,2657570995,9969267235]},
{"i":5390329,"b":"35: Noua - Terminal Gara","f":"Noua","t":"Terminal Gara","n":"35","c":"#b095a6","d":2,"s":[9969267235,611633517,611633516,611633522,611633524,611633525,611633519,3701356195,611633526,611633527,3701356196,3701356197,2683094124,2683092725,11801788121]},
{"i":5390330,"b":"36: Independentei - Livada Postei","f":"Independentei","t":"Livada Postei","n":"36","c":"#487a7b","d":1,"s":[3709437150,9171965584,9171965586,2657710288,3654024255,2657703726,3473944193,1923401922,2537929578,2552635273,2375041371]},
{"i":5390331,"b":"36: Livada Postei - Independentei","f":"Livada Postei","t":"Independentei","n":"36","c":"#487a7b","d":2,"s":[2375041371,2655859050,2679017945,1923401916,3652479574,3368804794,2657704969,2657706440,11681500791,11681500789,11671674450,11671674452,9171965589,3709437150]},
{"i":5410018,"b":"37: Craiter - Hidro A","f":"Craiter","t":"Hidro A","n":"37","c":"#764212","d":2,"s":[3713179295,3713179297,3652393629,3708943734,11801788119,2683090885,2683091443,267042578,8227802075]},
{"i":5410019,"b":"37: Hidro A - Craiter","f":"Hidro A","t":"Craiter","n":"37","c":"#764212","d":1,"s":[8227802075,2683094124,2683092725,2910431428,3652436632,2279493798,3713179296,3713179295]},
{"i":5390361,"b":"40: Lujerului - Terminal Gara","f":"Lujerului","t":"Terminal Gara","n":"40","c":"#ffc90d","d":2,"s":[3713232824,3713232813,3713233438,3708889572,3707831699,3707831700,3707831703,3707831701,3707831705,3707831709,3707831693,3709463826,3709463825,3709458524,3343778550,3343778546,3473944089,11801788125]},
{"i":5390360,"b":"40: Terminal Gara - Lujerului","f":"Terminal Gara","t":"Lujerului","n":"40","c":"#ffc90d","d":1,"s":[11801788125,3343778547,3343778551,3709449858,3709449854,9172134414,3709449855,3707786426,9171304342,3707786436,3707786432,3707786429,3707786430,3707786428,3707786427,3707786434,3831052316,3713232812,3713232824]},
{"i":5410088,"b":"41: Lujerului - Livada Postei","f":"Lujerului","t":"Livada Postei","n":"41","c":"#7bafd4","d":1,"s":[3713232824,3713232813,3713233438,3713233436,3713233431,3713233430,3713233426,9173869316,9173869313,3713233434,3713232808,3713232810,3707775824,3654024250,3708932940,3708932942,3708932943,3708932944,3708932941,2552635273,2375041372]},
{"i":5410087,"b":"41: Livada Postei - Lujerului","f":"Livada Postei","t":"Lujerului","n":"41","c":"#7bafd4","d":2,"s":[2375041372,2655859050,3709478534,3708912144,3708912143,3708912142,3708912141,3708932945,3708991157,3713232811,3713232809,3713233435,9173869312,9173869315,3713233425,3713233429,3713233432,3707786434,3831052316,3713232812,3713232824]},
{"i":14292150,"b":"50: Camera de Comert - Solomon","f":"Camera de Comert","t":"Solomon","n":"50","c":"#d0006e","d":2,"s":[3652393631,2537929578,2552635273,2375041368,3713320054,3713320055,2854669944,9174164592,9174164594,9174164596,9174164597,9174164604,9174164607,9174164610,9174164612]},
{"i":13329734,"b":"50: Solomon - Camera de Comert","f":"Solomon","t":"Camera de Comert","n":"50","c":"#d0006e","d":1,"s":[9174164612,10011185841,9174164608,9174164606,9174164599,9174164614,3713320058,3713320059,3713320056,3713320057,254344601,262148563,3652393631]},
{"i":13330002,"b":"52: Panselelor - Tocile","f":"Panselelor","t":"Tocile","n":"52","c":"#9ad3dc","d":1,"s":[9164803420,9164803422,2680755471,9274823440,611633519,3701409064,310840809,2537998221,2537995437,464246419,267042578,1923401922,2537929578,2552635273,3713320054,3713320055,2854669944,3713320058]},
{"i":13330003,"b":"52: Tocile - Panselelor","f":"Tocile","t":"Panselelor","n":"52","c":"#9ad3dc","d":2,"s":[3713320058,3713320059,3713320056,3713320057,254344601,262148563,8227802075,264864013,3653874651,2537994014,2537999895,2538006527,2538009452,3708904920,2680751671,9164803421,9164803420]},
{"i":13319272,"b":"53: Panselelor - Facultate Constructii","f":"Panselelor","t":"Facultate Constructii","n":"53","c":"#f82b3c","d":2,"s":[9164803420,9164803422,2680755471,9187345290,3701356195,611633526,611633527,3701356196,3701356197,2683094124,2683092725,1635108121,3368804794,9167564063,9167564061,9167564057]},
{"i":13319271,"b":"53: Facultate Constructii - Panselelor","f":"Facultate Constructii","t":"Panselelor","n":"53","c":"#f82b3c","d":1,"s":[9167564057,9167564059,2657703726,3473944089,2683090885,2683091443,267042578,8227802075,3640849654,310018170,3701356190,2657677035,3701356193,3713302845,9164803421,9164803420]},
{"i":14899833,"b":"54: Hidro A - Triaj","f":"Hidro A","t":"Triaj","n":"54","c":"#98999b","d":1,"s":[8227802075,3701356197,2683094124,2683092725,2910431428,3652436632,3652436626,3713264153,3713264157,3713264155,9164528578,9164528581,9164528582,9274917903]},
{"i":14899832,"b":"54: Triaj - Hidro A","f":"Triaj","t":"Hidro A","n":"54","c":"#98999b","d":2,"s":[9274917903,9164528578,9164528581,9164528582,3713264156,3713264158,3713264154,3652393629,3708943734,11801788119,2683090885,2683091443,267042578,8227802075]},
{"i":13326483,"b":"60: Silver Mountain - Telecabina","f":"Silver Mountain","t":"Telecabina","n":"60","c":"#6b99ba","d":1,"s":[9171470075,9473586964,10586930357,9171470095]},
{"i":13326484,"b":"60: Telecabina - Silver Mountain","f":"Telecabina","t":"Silver Mountain","n":"60","c":"#6b99ba","d":2,"s":[9171470095,9171470093,3708897319,9171470075]},
{"i":13688026,"b":"100: Terminal Gara - Telecabina","f":"Terminal Gara","t":"Telecabina","n":"100","c":"#0085ca","d":2,"s":[11801788120,3652393631,3652436629,3708889560,300099706,3708889559,3708889557,3708889571,10586930357,9171470095]},
{"i":13688025,"b":"100: Telecabina - Terminal Gara","f":"Telecabina","t":"Terminal Gara","n":"100","c":"#0085ca","d":1,"s":[9171470095,9171470093,3708897319,3708897315,3708897316,300099701,3708897317,2375041369,8227802075,11801788120]},
{"i":15962950,"b":"A1: Aeroportul Brasov - Terminal Gara","f":"Aeroportul Brasov","t":"Terminal Gara","n":"A1","c":"#ffffff","d":2,"s":[10964817435,3708889555,11014591991,9710744131,9171391885,3708974252,3708974250,3708974254,3708974253,3708974251,3701493888,9275068610,3701356185,3701356191,2375041369,8227802075,11801788120]},
{"i":17828248,"b":"A1: Terminal Gara - Scolii - Aeroportul Brasov","f":"Terminal Gara","t":"Aeroportul Brasov","n":"A1","c":"#ffffff","d":1,"s":[11801788120,3652393631,2375041368,12010293524,3654024249,9565453078,3713539337,3708962052,3708962054,3708962050,3708962051,9710744128,8926599968,3708889565,3707795549,10964817435]},
]
export default busses;