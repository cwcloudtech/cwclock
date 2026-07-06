# AI instruction 7

## Time tracking export

Members, administrators and owner can export the time tracking in CSV or PDF with a detailed export or a summary export compliant with clockify.

### Filters

The time export should be on a _report_ item in the sidebar with a graph icon it should propose a time selector like Grafana with two calendar and the following shortcuts:

* Today
* Yesterday
* This week
* Last week
* Past two weeks
* This month
* Last month
* Past two months
* This year
* Last year

Also other filters as autocomplete dropdown:

* Client
* Project
* Member

### Calculation rules

* If the record is "all day" set to true, the duration is set to the organization HoursPerDay
* Do not calculate with the VAT, it will be done in a future invoicing feature, just the sum of hours per member, devide by HoursPerDay the multiply per daily rate of each members

### Outputs

#### Frontend output

* [Detailed export](./assets/detailed-report-gui.png)
* [Summary export](./assets/summary-report-gui.png)

Notes:
* a member should not be able to see the calculated price from daily rate, only administrators and owner.
* ignore not existing exports like _Weekly_ I want only _Summary_ and _Detailed_
* in the detailed report, the record must be updatable or deletable (I can change the description or the member who make the record with drop down autocomplete)
* ignore the tags in the screenshots, it doesn't exists here

Of course adapt using the current UI/UX design, icons, etc and keep consistency. Those screenshots are just to illustrate the idea from a concurrent website.

#### CSV output

##### Summary

```csv
"Project","Client","Description","Time (h)","Time (decimal)","Amount (EUR)"
"INOVTECH-INOVTECH2022","inovshop","Resolve inovtech-server-api error (#275)","04:08:01","4.13","0.00"
"INTERNE-INTERNE","inovshop","Remove unused snapshots (#264)","00:34:30","0.58","0.00"
"OMNI-OMNI2022","inovshop","Update env variables (#274)","05:17:51","5.30","0.00"
```

##### Detailed

```csv
"Project","Client","Description","Task","User","Group","Email","Tags","Billable","Start Date","Start Time","End Date","End Time","Duration (h)","Duration (decimal)","Billable Rate (EUR)","Billable Amount (EUR)","Date of creation"
"INOVTECH-INOVTECH2022","inovtec","Resolve inovtech-server-api error (#275)","","B. Mohamed","","m-hedi@cwclock.me","","Yes","04/30/2026","02:44:33 PM","04/30/2026","03:19:30 PM","00:34:57","0.58","0.00","0.00","04/30/2026"
"MYPROJECT-MYPROJECT2022","inovtec","Update env variables (#274)","","B. Mohamed","","m-hedi@cwclock.me","","Yes","04/28/2026","10:22:52 AM","04/28/2026","01:06:29 PM","02:43:37","2.73","0.00","0.00","04/28/2026"
"INOVTECH-INOVTECH2022","inovtec","Resolve inovtech-server-api error (#275)","","B. Mohamed","","m-hedi@cwclock.me","","Yes","04/27/2026","08:48:09 PM","04/27/2026","10:42:26 PM","01:54:17","1.90","0.00","0.00","04/27/2026"
"MYPROJECT-MYPROJECT2022","inovtec","Update env variables (#274)","","B. Mohamed","","m-hedi@cwclock.me","","Yes","04/27/2026","03:56:13 PM","04/27/2026","04:29:13 PM","00:33:00","0.55","0.00","0.00","04/27/2026"
"INOVTECH-INOVTECH2022","inovtec","Resolve inovtech-server-api error (#275)","","B. Mohamed","","m-hedi@cwclock.me","","Yes","04/27/2026","02:17:16 PM","04/27/2026","03:56:03 PM","01:38:47","1.65","0.00","0.00","04/27/2026"
"MYPROJECT-MYPROJECT2022","inovtec","Update env variables (#274)","","B. Mohamed","","m-hedi@cwclock.me","","Yes","04/27/2026","12:16:01 PM","04/27/2026","02:17:15 PM","02:01:14","2.02","0.00","0.00","04/27/2026"
"INTERNE-INTERNE","inovtec","Remove unused snapshots (#264)","","B. Mohamed","","m-hedi@cwclock.me","","Yes","04/27/2026","11:27:53 AM","04/27/2026","12:02:23 PM","00:34:30","0.58","0.00","0.00","04/27/2026"
```

Notes: 
* on the billable column put always "Yes", it's for keeping compliance with clockify exports.
* same thing for empty columns like tags or groups
