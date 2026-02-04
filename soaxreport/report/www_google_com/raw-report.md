# Raw Data Analysis: ECH GREASE Connectivity (SOAX)

**Date:** February 02, 2026
**Target Domain:** `www.google.com`
**Analyzed File:** `soax-results-www_google_com-countries249.csv`

## Executive Summary

This report analyzed **866** valid ISP pairs. ECH GREASE does **not** appear to cause systematic connectivity breakage.

*   **Total ISP Pairs:** 866
*   **Potential Blocking:** 3 (0.35%)
*   **Avg Latency Impact:** 237.56 ms

## 1. Overall Connectivity Results

![Global Overview](global_overview.png)

**Figure 1: Global Connectivity Distribution.** This chart illustrates the health of tested ISP vantage points. "Healthy" represents successful connections with and without ECH. "Potential ECH Blocking" identifies cases where only the standard TLS succeeded. "Unreachable" indicates ISPs that failed both tests, likely due to proxy or local network issues unrelated to ECH.

## 2. Divergent Countries (Deep Dive)

![Problematic Countries](problematic_countries.png)

**Figure 2: Success Rate Divergence.** This chart only displays countries where the success rate of ECH GREASE differs from the control (No ECH). A significantly shorter red bar compared to the blue bar indicates a strong likelihood of ECH-specific interference in that country.

## 3. Performance Impact

![Latency Delta](latency_delta.png)

**Figure 3: Latency Delta Distribution.** The delta is calculated as `Handshake(GREASE) - Handshake(No ECH)`. Most values clustering around 0ms suggest that ECH GREASE does not introduce significant overhead when connections succeed.

## 4. Deep Dive: Potential Blocking

We detected **3** instances where ECH GREASE failed while the control succeeded. These cases warrant further investigation to distinguish between transient network errors and active blocking.

**Affected Countries:**
*   IN: 1 instance(s)
*   PT: 1 instance(s)
*   TL: 1 instance(s)

(See Appendix A for the full list of failures)

## 5. Limitations

*   **Transient Errors:** Single-pass testing cannot distinguish between flaky networks and deterministic blocking. Re-runs are required for confirmation.
*   **Proxy Stability:** Residential proxies (SOAX) can be inherently unstable or slow, which may contribute to timeouts independent of ECH.
*   **Sample Size:** The number of ISPs tested per country depends on SOAX's available pool at the time of testing.

## Appendix A: Detailed Failure List

| Country | ISP | No ECH Exit | GREASE Exit | Error Name |
| :--- | :--- | :--- | :--- | :--- |
| IN | jio | 0 | 56 | CURLE_RECV_ERROR |
| PT | digi portugal | 0 | 35 | CURLE_SSL_CONNECT_ERROR |
| TL | telkomcel | 0 | 56 | CURLE_RECV_ERROR |

## Appendix B: Significant Latency Increases (>500ms)

| Domain | Country | ISP | No ECH TLS (ms) | GREASE TLS (ms) | Delta (ms) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| www.google.com | IN | jio | 2448 | 24871 | 22423 |
| www.google.com | TL | telkomcel | 3455 | 21403 | 17948 |
| www.google.com | PK | multacom corporation | 0 | 16898 | 16898 |
| www.google.com | BI | ucom-wic | 14575 | 27739 | 13164 |
| www.google.com | TJ | cjsc babilon-mobile | 2414 | 8042 | 5628 |
| www.google.com | ML | sotelmabgp | 2140 | 7694 | 5554 |
| www.google.com | JP | softbank corp. | 2114 | 7217 | 5103 |
| www.google.com | KZ | tns-plus llp | 1661 | 5898 | 4237 |
| www.google.com | MW | airtel malawi | 2273 | 6074 | 3801 |
| www.google.com | FR | bouygues telecom | 4603 | 8349 | 3746 |
| www.google.com | PK | zong | 2064 | 5750 | 3686 |
| www.google.com | JP | k-opticom corporation | 1677 | 4618 | 2941 |
| www.google.com | CM | camtel | 1629 | 4472 | 2843 |
| www.google.com | GM | africell | 1977 | 4609 | 2632 |
| www.google.com | PE | fibra movistar | 1414 | 3830 | 2416 |
| www.google.com | NE | airtel niger | 2860 | 5249 | 2389 |
| www.google.com | CO | tigo colombia | 1301 | 3465 | 2164 |
| www.google.com | GH | airtel-ghana | 1607 | 3643 | 2036 |
| www.google.com | MW | tnm | 1707 | 3594 | 1887 |
| www.google.com | ZA | mtn business solutions | 2939 | 4824 | 1885 |
| www.google.com | AG | flow | 1508 | 3382 | 1874 |
| www.google.com | MQ | digicel antilles francaises guyane | 2390 | 4261 | 1871 |
| www.google.com | BF | orange burkina faso | 1117 | 2889 | 1772 |
| www.google.com | ZM | zamtel | 2034 | 3785 | 1751 |
| www.google.com | MQ | free mobile | 4051 | 5778 | 1727 |
| www.google.com | PE | bitel | 1370 | 3011 | 1641 |
| www.google.com | AF | afghan wireless communication company | 2183 | 3817 | 1634 |
| www.google.com | WS | vodafone samoa | 1925 | 3483 | 1558 |
| www.google.com | UY | movistar uruguay | 1202 | 2735 | 1533 |
| www.google.com | TL | viettel timor leste | 1705 | 3224 | 1519 |
| www.google.com | IN | airtel | 1421 | 2859 | 1438 |
| www.google.com | BF | onatel | 1391 | 2813 | 1422 |
| www.google.com | LB | mobile interim company 1 s.a.l. | 745 | 2159 | 1414 |
| www.google.com | MG | airtel madagascar | 1240 | 2650 | 1410 |
| www.google.com | NG | mtn nigeria | 2589 | 3974 | 1385 |
| www.google.com | CU | empresa de telecomunicaciones de cuba, s.a. | 1000 | 2368 | 1368 |
| www.google.com | NE | orange niger | 1786 | 3132 | 1346 |
| www.google.com | DE | play2go international | 1439 | 2769 | 1330 |
| www.google.com | RU | ekaterinburg-2000 | 1079 | 2390 | 1311 |
| www.google.com | ZM | airtel zambia | 2047 | 3358 | 1311 |
| www.google.com | NC | opt-nc | 1339 | 2634 | 1295 |
| www.google.com | FJ | digicel fiji | 1847 | 3139 | 1292 |
| www.google.com | BD | telenor | 1433 | 2716 | 1283 |
| www.google.com | FR | free sas | 1121 | 2377 | 1256 |
| www.google.com | AO | aas1 | 1355 | 2600 | 1245 |
| www.google.com | UG | mtn uganda | 1315 | 2554 | 1239 |
| www.google.com | UG | airtel uganda | 1513 | 2745 | 1232 |
| www.google.com | JP | japan communication | 2286 | 3507 | 1221 |
| www.google.com | SL | qcell | 1634 | 2854 | 1220 |
| www.google.com | HK | china mobile hong kong | 910 | 2128 | 1218 |
| www.google.com | BJ | moov benin | 1181 | 2398 | 1217 |
| www.google.com | YT | free reunion | 1414 | 2623 | 1209 |
| www.google.com | PG | vodafone png | 1593 | 2797 | 1204 |
| www.google.com | PY | claro argentina | 1337 | 2529 | 1192 |
| www.google.com | MN | mobicom corporation | 1679 | 2864 | 1185 |
| www.google.com | BT | druknet isp | 1331 | 2508 | 1177 |
| www.google.com | MN | univision | 1199 | 2374 | 1175 |
| www.google.com | SO | amtel | 1395 | 2566 | 1171 |
| www.google.com | KE | airtel rwanda | 1307 | 2477 | 1170 |
| www.google.com | MG | orange madagascar | 1503 | 2672 | 1169 |
| www.google.com | ET | safaricom | 1479 | 2646 | 1167 |
| www.google.com | JP | logiclinks | 1224 | 2379 | 1155 |
| www.google.com | MZ | movitel | 1224 | 2374 | 1150 |
| www.google.com | NZ | 2degrees | 2157 | 3306 | 1149 |
| www.google.com | NZ | spark new zealand | 1414 | 2546 | 1132 |
| www.google.com | SE | hi3g access ab | 550 | 1665 | 1115 |
| www.google.com | BO | entel bolivia | 1291 | 2396 | 1105 |
| www.google.com | TO | digicel tonga | 1510 | 2592 | 1082 |
| www.google.com | JP | rakuten mobile network | 1297 | 2377 | 1080 |
| www.google.com | TD | airtel chad | 1148 | 2228 | 1080 |
| www.google.com | BR | tim brasil | 1072 | 2152 | 1080 |
| www.google.com | JM | digicel jamaica | 883 | 1958 | 1075 |
| www.google.com | NA | mtc namibia | 3990 | 5041 | 1051 |
| www.google.com | KE | safaricom | 1280 | 2329 | 1049 |
| www.google.com | SD | mtn sudan | 1076 | 2122 | 1046 |
| www.google.com | TN | tunisie telecom | 1817 | 2853 | 1036 |
| www.google.com | NA | telecom namibia | 1912 | 2941 | 1029 |
| www.google.com | SG | starhub | 1070 | 2092 | 1022 |
| www.google.com | PG | digitec papua new guinea | 1365 | 2380 | 1015 |
| www.google.com | ES | yoigo | 985 | 1996 | 1011 |
| www.google.com | BR | surf telecom s.a. | 1236 | 2246 | 1010 |
| www.google.com | BR | claro brazil | 899 | 1908 | 1009 |
| www.google.com | BA | bh telecom d.d. sarajevo | 637 | 1646 | 1009 |
| www.google.com | KN | flow | 1133 | 2138 | 1005 |
| www.google.com | PH | globe telecom | 988 | 1979 | 991 |
| www.google.com | LK | mobitel | 1003 | 1991 | 988 |
| www.google.com | TW | chunghwa telecom | 996 | 1982 | 986 |
| www.google.com | TZ | airtel tanzania | 1256 | 2225 | 969 |
| www.google.com | BD | robi | 995 | 1948 | 953 |
| www.google.com | LS | econet telecom lesotho | 1685 | 2632 | 947 |
| www.google.com | MG | telecom-malagasy | 1498 | 2427 | 929 |
| www.google.com | AU | telstra internet | 1635 | 2563 | 928 |
| www.google.com | MM | atom myanmar | 1021 | 1941 | 920 |
| www.google.com | DK | tdc net | 1545 | 2459 | 914 |
| www.google.com | EG | vodafone egypt | 1336 | 2248 | 912 |
| www.google.com | MG | gulfsat-madagascar | 1514 | 2426 | 912 |
| www.google.com | CG | airtel congo | 1903 | 2810 | 907 |
| www.google.com | GH | mtn ghana | 1063 | 1968 | 905 |
| www.google.com | IR | mtn irancell | 1741 | 2644 | 903 |
| www.google.com | CD | airtel drc | 1339 | 2236 | 897 |
| www.google.com | SI | a1 slovenija | 908 | 1801 | 893 |
| www.google.com | RU | s.u.e. dpr republic operator of networks | 649 | 1537 | 888 |
| www.google.com | TZ | vodacom tanzania | 1019 | 1901 | 882 |
| www.google.com | LS | vodacom-lesotho | 1512 | 2392 | 880 |
| www.google.com | SZ | swazimtn-ltd | 1013 | 1891 | 878 |
| www.google.com | JP | arteria networks corporation | 2613 | 3485 | 872 |
| www.google.com | BD | grameenphone | 963 | 1835 | 872 |
| www.google.com | ET | ethiopian telecommunication corporation | 935 | 1806 | 871 |
| www.google.com | JP | internet initiative japan | 1108 | 1972 | 864 |
| www.google.com | BR | brisanet | 1075 | 1938 | 863 |
| www.google.com | VN | vietnamobile telecommunications joint stock compan | 1091 | 1946 | 855 |
| www.google.com | IQ | asiacell communications pjsc | 1226 | 2079 | 853 |
| www.google.com | TN | ooredoo tunisia | 507 | 1359 | 852 |
| www.google.com | ZA | mtn sa mobile | 1129 | 1979 | 850 |
| www.google.com | BI | viettel burundi | 1286 | 2133 | 847 |
| www.google.com | PE | movistar | 1900 | 2746 | 846 |
| www.google.com | TW | fareastone | 1172 | 2018 | 846 |
| www.google.com | RU | jsc vainah telecom | 864 | 1709 | 845 |
| www.google.com | LR | orange liberia | 1066 | 1906 | 840 |
| www.google.com | LK | hutch sri lanka | 905 | 1723 | 818 |
| www.google.com | MU | mauritius telecom | 1584 | 2393 | 809 |
| www.google.com | IN | vodafone idea | 2450 | 3248 | 798 |
| www.google.com | GA | gabon-telecom | 1454 | 2248 | 794 |
| www.google.com | MN | g-mobile corporation | 860 | 1650 | 790 |
| www.google.com | TH | true mobile | 820 | 1605 | 785 |
| www.google.com | ZA | telkom internet | 1519 | 2300 | 781 |
| www.google.com | AU | vocus | 6959 | 7737 | 778 |
| www.google.com | TJ | closed joint stock company tt mobile | 758 | 1534 | 776 |
| www.google.com | GU | lumen | 855 | 1627 | 772 |
| www.google.com | JP | ntt docomo | 1290 | 2056 | 766 |
| www.google.com | HR | hrvatski telekom | 789 | 1547 | 758 |
| www.google.com | KH | metfone | 845 | 1603 | 758 |
| www.google.com | AL | vodafone albania | 665 | 1422 | 757 |
| www.google.com | JO | umniah | 1573 | 2329 | 756 |
| www.google.com | GY | u mobile cellular inc. | 1511 | 2262 | 751 |
| www.google.com | IQ | seven net | 902 | 1648 | 746 |
| www.google.com | EC | conecel | 865 | 1605 | 740 |
| www.google.com | JP | au one net | 1036 | 1773 | 737 |
| www.google.com | ID | xl axiata | 1221 | 1956 | 735 |
| www.google.com | SK | slovak telekom | 711 | 1444 | 733 |
| www.google.com | AR | claro argentina | 749 | 1480 | 731 |
| www.google.com | JE | jtglobal | 785 | 1514 | 729 |
| www.google.com | RU | beeline | 1148 | 1874 | 726 |
| www.google.com | CM | mtn cameroon | 808 | 1533 | 725 |
| www.google.com | TG | atlantique telecom | 1180 | 1904 | 724 |
| www.google.com | TW | twn broadband | 1537 | 2259 | 722 |
| www.google.com | HK | csl mobile | 844 | 1565 | 721 |
| www.google.com | KH | flash broadband pvt. ltd. | 725 | 1441 | 716 |
| www.google.com | IL | pelephone | 755 | 1469 | 714 |
| www.google.com | UZ | unitel llc | 937 | 1649 | 712 |
| www.google.com | VN | viettel group | 827 | 1537 | 710 |
| www.google.com | BD | banglalink digital communications ltd. | 1416 | 2119 | 703 |
| www.google.com | US | verizon wireless | 1010 | 1713 | 703 |
| www.google.com | SO | hormuud | 1186 | 1886 | 700 |
| www.google.com | ZA | cell c | 1043 | 1731 | 688 |
| www.google.com | RU | tele2 russia | 908 | 1591 | 683 |
| www.google.com | TH | ais eds | 681 | 1363 | 682 |
| www.google.com | IR | mobile communication company of iran | 1756 | 2437 | 681 |
| www.google.com | IS | nova hf | 1654 | 2332 | 678 |
| www.google.com | PF | vini | 787 | 1464 | 677 |
| www.google.com | SA | stc saudi | 855 | 1532 | 677 |
| www.google.com | ZW | netone-cellular | 2516 | 3185 | 669 |
| www.google.com | TH | ais mobile | 1140 | 1809 | 669 |
| www.google.com | IT | tiscali | 533 | 1198 | 665 |
| www.google.com | RU | pjsc megafon | 723 | 1387 | 664 |
| www.google.com | MZ | vodacom mozambique | 1221 | 1881 | 660 |
| www.google.com | RU | mts pjsc | 784 | 1443 | 659 |
| www.google.com | GB | sure south atlantic | 1103 | 1759 | 656 |
| www.google.com | BF | telecel-faso | 1156 | 1805 | 649 |
| www.google.com | TZ | mic tanzania | 1800 | 2445 | 645 |
| www.google.com | CY | cablenet communication systems | 1074 | 1716 | 642 |
| www.google.com | IQ | al atheer telecommunication-iraq co. incorporated | 796 | 1434 | 638 |
| www.google.com | UY | claro uruguay | 549 | 1183 | 634 |
| www.google.com | GB | sparks communications | 458 | 1090 | 632 |
| www.google.com | PY | tigo paraguay | 898 | 1530 | 632 |
| www.google.com | SA | zain saudi arabia | 846 | 1473 | 627 |
| www.google.com | ES | vodafone spain | 575 | 1196 | 621 |
| www.google.com | NG | airtel networks limited | 1002 | 1622 | 620 |
| www.google.com | MC | monaco telecom | 647 | 1264 | 617 |
| www.google.com | BH | stc bahrain | 637 | 1253 | 616 |
| www.google.com | ES | avatel telecom | 1280 | 1887 | 607 |
| www.google.com | ZA | vodacom | 1052 | 1659 | 607 |
| www.google.com | GB | lycamobile | 1101 | 1705 | 604 |
| www.google.com | JP | open computer network | 1259 | 1859 | 600 |
| www.google.com | EG | telecom egypt | 1192 | 1789 | 597 |
| www.google.com | CV | tmais | 1035 | 1626 | 591 |
| www.google.com | GP | outremer telecom | 618 | 1208 | 590 |
| www.google.com | LR | lonestar | 926 | 1516 | 590 |
| www.google.com | GW | mtn-bissau | 1099 | 1687 | 588 |
| www.google.com | TT | telecommunication services of trinidad and tobago | 1290 | 1878 | 588 |
| www.google.com | PK | paknet merged into ptcl | 1072 | 1660 | 588 |
| www.google.com | BW | botswana telecommunications corporation | 1205 | 1787 | 582 |
| www.google.com | CM | orange cameroun | 1030 | 1607 | 577 |
| www.google.com | AZ | azercell telecom | 869 | 1446 | 577 |
| www.google.com | RE | zeop | 1750 | 2326 | 576 |
| www.google.com | IT | vodafone italia | 602 | 1173 | 571 |
| www.google.com | AR | personal | 744 | 1309 | 565 |
| www.google.com | KZ | jusan mobile jsc | 646 | 1202 | 556 |
| www.google.com | DZ | algerie telecom | 495 | 1047 | 552 |
| www.google.com | KR | sk telecom | 1321 | 1872 | 551 |
| www.google.com | UY | antel uruguay | 729 | 1279 | 550 |
| www.google.com | RU | rostelecom | 799 | 1347 | 548 |
| www.google.com | GY | e-networks inc | 972 | 1519 | 547 |
| www.google.com | GT | claro guatemala | 1211 | 1751 | 540 |
| www.google.com | RS | a1 srbija d.o.o | 593 | 1132 | 539 |
| www.google.com | UZ | coscom liability company | 2234 | 2771 | 537 |
| www.google.com | KW | zain kuwait | 1064 | 1599 | 535 |
| www.google.com | MX | at&t mexico | 835 | 1369 | 534 |
| www.google.com | RU | novokuznetsk telecom | 514 | 1044 | 530 |
| www.google.com | TR | turk telekom | 700 | 1229 | 529 |
| www.google.com | MK | a1 makedonija | 652 | 1175 | 523 |
| www.google.com | NP | nepal telecom | 1056 | 1576 | 520 |
| www.google.com | AG | digicel | 516 | 1035 | 519 |
| www.google.com | IL | wecom mobile | 1044 | 1563 | 519 |
| www.google.com | SN | sudatel-senegal | 924 | 1442 | 518 |
| www.google.com | MV | ooredoo maldives | 1248 | 1764 | 516 |
| www.google.com | JP | ntt communications corporation | 1158 | 1667 | 509 |
| www.google.com | SA | zain kuwait | 837 | 1345 | 508 |
| www.google.com | EE | tele2 estonia | 654 | 1160 | 506 |
| www.google.com | PL | orange polska | 676 | 1179 | 503 |
| www.google.com | PE | movistar peru | 1059 | 1562 | 503 |

## Appendix C: Data Anomalies (Unpaired or Duplicate Rows)

_No data anomalies found._
