---
title: "Detección de Desperdicio"
weight: 40
type: docs
prev: /docs/cost-analytics
next: /docs/reporting
sidebar:
  collapsed: false
---

El motor de **Detección de Desperdicio** es el módulo de diagnóstico central de **AWS Doctor**. Escanea su cuenta en busca de recursos "zombie": activos que están activos y facturando pero que no proporcionan ningún valor a su negocio.

## Cómo Ejecutar
Utilice el subcomando `waste` para activar un escaneo completo en todos los servicios soportados:

```bash
aws-doctor waste --region us-east-1
```

![Escaneo de Detección de Desperdicio](/images/demo/waste.gif)

### Escaneo Selectivo
Si sólo desea escanear servicios de AWS específicos, puede pasarlos como argumentos al subcomando. Esto es útil para una ejecución más rápida o para limpiezas específicas.

| Argumento | Servicio |
| :--- | :--- |
| `ec2` | Instancias EC2, volúmenes EBS, snapshots, key pairs, AMIs. |
| `elb` | Load Balancers Application y Network. |
| `s3` | Buckets S3 y cargas multiparte. |
| `rds` | Instancias RDS y snapshots. |
| `lambda` | Detección de memoria sobre-provisionada en Lambda. |
| `vpc` | NAT Gateways y recursos de VPC inactivos. |
| `cloudwatch` | Grupos de logs de CloudWatch sin políticas de retención. |
| `sagemaker` | Detección de endpoints de SageMaker inactivos (cero invocaciones en 14 días). |

```bash
# Ejemplo: Escanear solo recursos de EC2 y SageMaker
aws-doctor waste ec2 sagemaker
```

### Flags de Configuración

Los subcomandos `waste` y `report waste` soportan flags específicos para ajustar la lógica de detección:

| Flag | Por Defecto | Descripción |
| :--- | :--- | :--- |
| `--lambda-memory-threshold` | `10` | Umbral de utilización de memoria (%) por debajo del cual las funciones Lambda se marcan como sobre-provisionadas. |

## Categorías de Detección

Agrupamos el desperdicio en categorías principales de infraestructura:

{{< hextra/feature-grid cols="2" >}}
  {{< hextra/feature-card
    icon="server"
    title="Cómputo y EBS"
    link="compute/"
    subtitle="Instancias EC2 detenidas por >30 días, volúmenes huérfanos, snapshots obsoletos y RIs vencidas."
  >}}
  {{< hextra/feature-card
    icon="code"
    title="Lambda"
    link="compute/#aws-lambda"
    subtitle="Detección de memoria sobre-provisionada y recomendaciones de ajuste de tamaño."
  >}}
  {{< hextra/feature-card
    icon="database"
    title="Bases de Datos"
    link="databases/"
    subtitle="Instancias RDS detenidas, snapshots manuales de más de 30 días e instancias de base de datos inactivas."
  >}}
  {{< hextra/feature-card
    icon="archive"
    title="Almacenamiento y Logs"
    link="storage/"
    subtitle="Buckets sin políticas de ciclo de vida, cargas multipartes incompletas y grupos de logs sin retención."
  >}}
  {{< hextra/feature-card
    icon="share"
    title="Redes"
    link="networking/"
    subtitle="IPs Elásticas sin asociar, Load Balancers inactivos y NAT Gateways inactivos."
  >}}
  {{< hextra/feature-card
    icon="chip"
    title="Machine Learning"
    link="machine-learning/"
    subtitle="Endpoints de SageMaker activos con cero invocaciones en los últimos 14 días."
  >}}
{{< /hextra/feature-grid >}}

---

## ¿Por qué automatizar esto?
En las organizaciones grandes, los desarrolladores a menudo crean recursos temporales (probando una AMI, lanzando una EIP de prueba) y se olvidan de eliminarlos. Con el tiempo, estos pequeños cargos se agregan en miles de dólares de "deuda de infraestructura".

**AWS Doctor** hace que sea trivial realizar un chequeo semanal y mantener su cuenta optimizada.
