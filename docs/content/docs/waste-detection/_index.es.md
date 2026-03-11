---
title: "Detección de Desperdicio"
weight: 40
type: docs
prev: /docs/cost-analytics
next: /docs/automation
sidebar:
  collapsed: false
---

El motor de **Detección de Desperdicio** es el módulo de diagnóstico central de **AWS Doctor**. Escanea su cuenta en busca de recursos "zombie": activos que están activos y facturando pero que no proporcionan ningún valor a su negocio.

## Cómo Ejecutar
Utilice el flag `--waste` para activar un escaneo completo en todos los servicios soportados:

```bash
aws-doctor --waste --region us-east-1
```

![Escaneo de Detección de Desperdicio](/images/demo/waste.gif)

### Escaneo Selectivo
Si sólo desea escanear servicios de AWS específicos, puede pasar una lista separada por comas directamente al flag. Esto es útil para una ejecución más rápida o para limpiezas específicas.

Los filtros soportados actualmente son `ec2`, `s3`, `elb`, y `cloudwatch`.

```bash
# Ejemplo: Escanear solo recursos de EC2 y S3
aws-doctor --waste ec2,s3 --region us-east-1

# Ejemplo: Escanear solo CloudWatch Logs
aws-doctor --waste cloudwatch --region us-east-1
```

## Categorías de Detección

Agrupamos el desperdicio en tres categorías principales de infraestructura:

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    icon="server"
    title="Cómputo y EBS"
    link="compute/"
    subtitle="Instancias detenidas por >30 días, volúmenes huérfanos, snapshots obsoletos y RIs vencidas."
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
    subtitle="IPs Elásticas sin asociar y Load Balancers sin objetivos saludables."
  >}}
{{< /hextra/feature-grid >}}

---

## ¿Por qué automatizar esto?
En las organizaciones grandes, los desarrolladores a menudo crean recursos temporales (probando una AMI, lanzando una EIP de prueba) y se olvidan de eliminarlos. Con el tiempo, estos pequeños cargos se agregan en miles de dólares de "deuda de infraestructura".

**AWS Doctor** hace que sea trivial realizar un chequeo semanal y mantener su cuenta optimizada.
