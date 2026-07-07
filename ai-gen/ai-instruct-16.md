# AI instruction 16

## Summary export

In the bar chart printed in the pdf export summary, we miss the days below the bar.

## Importing data from clockify

An admin or owner of organization can import detailed CSV format:

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

It will create time entry if there's no entry already matching.

It will create the client and project with a random color if it doesn't exists.

It will also create the user as disabled if it doesn't exists (parsing the `User` column of the csv splitting with space in this order : `Lastname Firstname`).
