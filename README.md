# go_queue_example

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

Note: After you click the button you need to `heroku ps:scale worker=1 -a <app name>` or the [dashboard](https://heroku.com) equivalent.

Go based Queue / Background Worker API only example.

After app setup you can test with the following commands:

In one terminal run the following...

```term
heroku logs --tail -a <app name>
```

In a different terminal run the following...

```term
curl -XPOST "https://<app name>.herokuapp.com/index" -d '{"url": "http://google.com"}'
```

And you should see something like the following scroll by in the first terminal...

```term
2015-06-23T18:29:35.663096+00:00 heroku[router]: at=info method=POST path="/index" host=<app name>.herokuapp.com request_id=84f9d369-7d6e-4313-8f16-9db9bb7ed251 fwd="76.115.27.201" dyno=web.1 connect=19ms service=31ms status=202 bytes=141
2015-06-23T18:29:35.623878+00:00 app[web.1]: [negroni] Started POST /index
2015-06-23T18:29:35.644483+00:00 app[web.1]: [negroni] Completed 202 Accepted in 20.586125ms
2015-06-23T18:29:37.750543+00:00 app[worker.1]: time="2015-06-23T18:29:37Z" level=info msg="Processing IndexRequest! (not really)" IndexRequest={http://google.com}
2015-06-23T18:29:37.753021+00:00 app[worker.1]: 2015/06/23 18:29:37 event=job_worked job_id=1 job_type=IndexRequests
```

This shows the web process getting the request to index a url (http://google.com) and then the worker picking up the raw job and "processing" it.
