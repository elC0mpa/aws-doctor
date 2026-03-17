---
title: "Análisis de Costos"
description: "Descubra los dos flujos de trabajo principales de análisis de costos: Comparativo y de Tendencia."
weight: 30
type: docs
prev: /docs/usage
next: /docs/waste-detection
---

**AWS Doctor** se centra en el análisis de costos contextual. En lugar de mostrar solo números brutos, le ayuda a entender cómo evoluciona su gasto.

## Flujo Comparativo

Para obtener una comparación paralela de su gasto, utilice el subcomando `cost`:

```bash
aws-doctor cost
```

Esto activa el **Flujo Comparativo**, que incluye un desglose por servicio (EC2, S3, etc.) para ayudarle a identificar los impulsores de costos específicos.

![Análisis Comparativo de Costos](/images/demo/basic.gif)

### Justicia en la Comparación
Las comparaciones de facturación tradicionales (como mes a mes) suelen ser engañosas. Comparar el 15 de octubre con el mes completo de septiembre siempre parecerá un "ahorro", incluso si está gastando más.

**AWS Doctor** soluciona esto comparando ventanas de tiempo idénticas:
- **Periodo Actual**: 1er día del mes → Hoy.
- **Periodo Anterior**: 1er día del mes anterior → Día idéntico del mes pasado.

*Ejemplo: Si hoy es 15 de octubre, compara del 1 al 15 de octubre con el 1 al 15 de septiembre.*

{{< callout type="warning" >}}
**1er Día del Mes**: Esta función no está disponible el primer día del mes. AWS Cost Explorer requiere un rango mínimo de 24 horas donde la fecha de inicio sea estrictamente anterior a la fecha de finalización.
{{< /callout >}}

---

## Análisis de Tendencias de 6 Meses

Para detectar patrones de crecimiento a largo plazo o cambios arquitectónicos repentinos, utilice el subcomando `trend`:

```bash
aws-doctor trend
```

![Análisis de Tendencias de 6 Meses](/images/demo/trend.gif)

### Filtrado por Servicio
Ahora puede centrar su análisis de tendencias en servicios específicos de AWS pasándolos como argumentos al subcomando `trend`. Esto es útil para monitorear la evolución del costo de un solo servicio o un grupo de servicios.

```bash
# Analizar tendencia de costo solo para EC2 y S3
aws-doctor trend ec2 s3
```

Puede pasar múltiples servicios separados por espacios o comas.

### Abreviaturas de Servicios Disponibles
**AWS Doctor** utiliza abreviaturas intuitivas para el filtrado. Cada abreviatura se asigna a su nombre oficial de servicio de AWS Cost Explorer:

| Abreviatura | Servicio AWS | Abreviatura | Servicio AWS |
| :--- | :--- | :--- | :--- |
| `ec2` | EC2 Compute | `s3` | S3 Storage |
| `rds` | RDS Database | `lambda` | AWS Lambda |
| `dynamodb` | DynamoDB | `eks` | EKS (Kubernetes) |
| `ecs` | ECS (Containers) | `elb` | Elastic Load Balancing |
| `vpc` | Virtual Private Cloud | `route53` | Route 53 |
| `apigateway` | API Gateway | `cloudfront` | CloudFront |
| `cloudwatch` | CloudWatch | `elasticache` | ElastiCache |
| `redshift` | Redshift | `savingsplans` | Savings Plans |
| `glue` | AWS Glue | `kinesis` | Kinesis |
| `firehose` | Kinesis Firehose | `quicksight` | QuickSight |
| `waf` | AWS WAF | `backup` | AWS Backup |
| `stepfunctions`| Step Functions | `kms` | KMS |
| `secretsmanager`| Secrets Manager | `ssm` | Systems Manager |
| `location` | Location Service | `ecr` | ECR (Registry) |

### Qué muestra:
- Un gráfico de barras ANSI de alta fidelidad en su terminal.
- Costos totales mensuales para los últimos 6 ciclos de facturación completados.
- Indicadores claros de si su gasto se está acelerando o estabilizando.
