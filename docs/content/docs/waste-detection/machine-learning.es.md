---
title: "Machine Learning"
description: "Detecte endpoints de AWS SageMaker inactivos que están facturando pero no se están utilizando."
weight: 60
type: docs
prev: /docs/waste-detection/networking
next: /docs/reporting
---

Los recursos de Machine Learning, particularmente los endpoints de **Amazon SageMaker**, suelen estar entre los componentes más caros de una arquitectura de AWS. Es común que los científicos de datos y desarrolladores desplieguen endpoints para pruebas o validación y se olviden de darlos de baja una vez que el trabajo ha finalizado.

**AWS Doctor** le ayuda a identificar estos endpoints inactivos para evitar costos innecesarios.

## Endpoints de SageMaker

Actualmente, la herramienta audita los endpoints en tiempo real de SageMaker en su cuenta.

### Lógica de Detección
Un endpoint se marca como **inactivo** si cumple con los siguientes criterios:
1.  **Estado**: El endpoint debe estar en el estado `InService`.
2.  **Invocaciones**: La métrica de CloudWatch `SageMakerVariantInvocations` (Sum) debe ser **0** para todas las variantes asociadas con el endpoint durante los últimos **14 días**.

### ¿Por qué 14 días?
Utilizamos una ventana de observación de 14 días para tener en cuenta los modelos que podrían utilizarse para procesamiento quincenal o tareas programadas, asegurándonos de no marcar recursos que aún forman parte de un flujo de trabajo de producción regular (aunque poco frecuente).

## Cómo Ejecutar
Para ejecutar la detección de desperdicio de SageMaker de forma individual:

```bash
aws-doctor waste sagemaker
```

## Remediación
Si un endpoint se marca como inactivo:
1.  **Verificar**: Compruebe si el modelo aún es necesario para futuras solicitudes o si formó parte de un experimento completado.
2.  **Eliminar**: Si el endpoint ya no es necesario, elimínelo a través de la consola de AWS o la CLI. Recuerde que la facturación de SageMaker se basa en el tipo de instancia y la duración en que el endpoint está activo, independientemente de si procesa solicitudes.
3.  **Actualizar**: Si necesita el modelo ocasionalmente pero no las 24 horas del día, los 7 días de la semana, considere utilizar **SageMaker Serverless Inference** o **Inference Pipelines** para reducir costos.
