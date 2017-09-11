# Mitose
Easy Kubernetes autoscaler controller.
![mitose](http://biologianet.uol.com.br/upload/conteudo/images/na-mitose-uma-celula-mae-origina-duas-celulas-filhas-com-mesmo-numero-cromossomos-5964d9200973d.jpg)

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
  "key": "XXXX",
  "secret": "XXXX",
  "region": "us-east-1",
  "queue_urls": ["https://sqs.us-east-1.amazonaws.com/XXXXXXX/XXXXXXX"],
  "msgs_per_pod": 2
}
```
Save that content as `target.json` file and create a configmap
using the `kubectl create configmap` command, f.ex:
```shell
$ kubectl create configmap config --from-file=target.json --namespace=mitose
```

## TODO
- Tests
- Admin to _CRUD_ the configs.
- Kubernetes Deploy Yaml.
