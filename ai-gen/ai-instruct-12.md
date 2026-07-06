# AI instruction 12

## Update client or project

An administrator or owner can edit a client or project, add an edit button and edit form to the client list.

Update also the backend with a `PUT` endpoint if it doesn't exists.

## Stamp

Add stamp (or "cachet" in French) stored in base64 in organization's data.

## Name of downloaded exports

# Example of the exports

Pattern:

```
CWClock_Time_Report_{type}_{startDate}-{endDate}.{extension}
```

Examples:

```
CWClock_Time_Report_Summary_04_27_2026-05_03_2026.csv
CWClock_Time_Report_Detailed_04_27_2026-05_03_2026.csv
CWClock_Time_Report_Summary_04_27_2026-05_03_2026.pdf
CWClock_Time_Report_Detailed_04_27_2026-05_03_2026.pdf
```
