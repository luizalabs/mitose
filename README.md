# Mitose
Easy Kubernetes autoscaler controller.
![mitose](http://biologianet.uol.com.br/upload/conteudo/images/na-mitose-uma-celula-mae-origina-duas-celulas-filhas-com-mesmo-numero-cromossomos-5964d9200973d.jpg)

## Install
To install mitose in your k8s cluster just run:
```
$ kubectl create -f mitose-app.yaml
```
> We recommended you to use a diferent namespace

## Controllers Configuration
The Mitose controllers are configured by kubernetes [configmaps](https://kubernetes.io/docs/tasks/configure-pod-container/configmap/).
Each entry on configmap represents a deployment to watch.
The configuration data format is `json` with those fields:

Field | Description
----- | -----------
namespace | namespace of deployment
deployment | deployment name
type | type of controller
max | maximum number of replicas
min | minimum number of replicas
scale\_method | method of autoscaling (by editing `HPA` or editing `DEPLOY`)
interval | controller running interval (e.g. `1m`)
active | if this controller is active

> Those fields are comom for each controller type.

You don't need to restart mitose when you change a configmap,
because mitose will rebuild its controllers on each configmap change.

### SQS Queue Size Controller
There is a mitose controller bases on AWS SQS queue size.
The specifics configuration fields are:

Field | Description
----- | -----------
key | aws credential key
secret | aws credential secret
region | aws region
queue\_urls | a list of the complete endopoints of the queues
msgs\_per\_pod | the desired number of msgs in queue per replica

## Example
To configure a controller based on SQS queue size use the follow example:
```json
{
  "namespace": "target",
  "deployment": "target",
  "type": "sqs",
  "interval": "1m",
  "scale_method": "DEPLOY",
  "max": 5,
  "min": 1,
  "active": true,
  "key": "XXXX",
  "secret": "XXXX",
  "region": "us-east-1",
  "queue_urls": ["https://sqs.us-east-1.amazonaws.com/XXXXXXX/XXXXXXX"],
  "msgs_per_pod": 2
}
```

Or to configure a controller based on GCP's Pub/Sub:

```json
{
  "namespace": "target",
  "deployment": "target",
  "type": "pubsub",
  "interval": "1m",
  "scale_method": "DEPLOY",
  "max": 5,
  "min": 1,
  "active": true,
  "google_application_credentials": "XXXX",
  "region": "us-east1",
  "subscription_ids": ["mysub"],
  "project": "my-gcp-project",
  "msgs_per_pod": 2
}
```

`google_application_credentials` should be the location of a `credentials.json` file provided by GCP.

Save that content as `target.json` file and create a configmap
using the `kubectl create configmap` command, f.ex:
```shell
$ kubectl create configmap config --from-file=target.json --namespace=mitose
```
## Prometheus metrics handler configuration
To expose mitose metrics to prometheus you need to expose a service to deploy
```
$ kubectl expose deployment mitose --type=ClusterIP --port=5000 --namespace mitose
```
and add the annotation `prometheus.io/scrape: "true"` on this service.
Mitose will start a http server on the port configured by environment variable `$PORT`.

> If you deployed mitose using the `mitose-app.yaml` file, you don't need that.

## TODO
- Tests
- Admin to _CRUD_ the configs.
