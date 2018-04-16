## Problem:

In the error logs, I'm seeing alot of timeouts, related to projects and other api-calls.

### Explanation: 

In a hub that is either busy, slow, or undergoing a period of slowness due to KB dependencies, you might see errors like this.

```
time="2018-04-16T17:48:01Z" level=error msg="Error getting HTTP Response: Get https://35.192.210.92/api/projects/a1206acc-6792-4717-80f6-b9c37f46c891/versions/eefacf48-a003-4d21-b5b7-8dc0cf0222c7/risk-profile: net/http: request canceled (Client.Timeout exceeded while awaiting headers)."
time="2018-04-16T17:48:01Z" level=error msg="Error trying to retrieve a project version risk profile: Get https://35.192.210.92/api/projects/a1206acc-6792-4717-80f6-b9c37f46c891/versions/eefacf48-a003-4d21-b5b7-8dc0cf0222c7/risk-profile: net/http: request canceled (Client.Timeout exceeded while awaiting headers)."
time="2018-04-16T17:48:01Z" level=error msg="error fetching project version risk profile: Get https://35.192.210.92/api/projects/a1206acc-6792-4717-80f6-b9c37f46c891/versions/eefacf48-a003-4d21-b5b7-8dc0cf0222c7/risk-profile: net/http: request canceled (Client.Timeout exceeded while awaiting headers)"
time="2018-04-16T17:48:01Z" level=error msg="error checking hub for completed scan for sha ddc1fe8358087d56a2c7c421c6b3cbbf770f05162471e0cb4b62875105821573: Get https://35.192.210.92/api/projects/a1206acc-6792-4717-80f6-b9c37f46c891/versions/eefacf48-a003-4d21-b5b7-8dc0cf0222c7/risk-profile: net/http: request canceled (Client.Timeout exceeded while awaiting headers)"e2d681b6e3b7e8eedf2fbb288c3e6587db6fd2b7a4aa55dd3a8ab032094dfa8c: Get https://35.192.210.92/api/projects/f66cab69-64e5-4250-8425-09674f0d779a/versions/c940c92e-e85b-486d-b4c9-6ef733665095/risk-profile: net/http: request canceled 
```
This is generally okay, as long as progress is being made.  To check , you can check that `Every 2.0s: curl http://opssight-perceptor:3001/metrics ` shows a "ScanStatusComplete" value that is increasing over time.  

*Solution*

If, for several minutes, you see no scan status completed, do the following:

- Ensure your postgres server is performing well, and has at least 16GB of memory and 4 cores as well.
- Ensure your webapp has at 32GB of memory and 4 cores.
- Restart your hub webapp.  Then restart your perceptor, perceiver containers.
