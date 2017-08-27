# Mitose
Easy Kubernetes autoscaler controler.
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

> Those fields are comom for each controller type.

### SQS Queue Size Controller
There is a mitose controller bases on AWS SQS queue size.
The specifics configuration fields are:

Field | Description
----- | -----------
queue_url | the complete endopoint of the queue
msgs_per_pod | the desired number of msgs in queue per replica

## Example
To configure a controller based on SQS queue size use the follow example:
```json
{
  "namespace": "target",
  "deployment": "target",
  "type": "sqs",
  "max": 5,
  "min": 1,
  "queue_url": "https://sqs.us-east-1.amazonaws.com/XXXXXXX/XXXXXXX",
  "msgs_per_pod": 2
}
```
Save that content as `target.json` file and create a configmap
using the `kubectl create configmap` command, f.ex:
```shell
$ kubectl create configmap config --from-file=target.json --namespace=mitose
```
