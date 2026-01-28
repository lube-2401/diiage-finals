# Backend Application

Application backend HTTP écrite en Go qui expose une API REST et interagit avec l'API Kubernetes.

## Fonctionnalités

L'application expose plusieurs endpoints HTTP :

- **`GET /`** : Affiche le namespace dans lequel l'application s'exécute
- **`GET /config`** : Lit et retourne le contenu complet du ConfigMap `app-config` depuis le namespace courant
- **`GET /pods`** : Liste tous les pods présents dans le cluster (tous namespaces confondus) au format JSON
- **`GET /health`** : Endpoint de health check qui retourne "OK"
- **`GET /metrics`** : Expose les métriques Prometheus de l'application

## Métriques Prometheus

L'application expose les métriques suivantes :

- `http_requests_total` : Compteur du nombre total de requêtes HTTP par chemin et code de statut
- `configmap_read_total` : Compteur du nombre de lectures de ConfigMap

## Configuration

L'application utilise les variables d'environnement suivantes :

- `POD_NAMESPACE` : Namespace Kubernetes dans lequel l'application s'exécute (défaut: `default`)
- `PORT` : Port d'écoute du serveur HTTP (défaut: `8080`)

## Prérequis

L'application doit s'exécuter dans un cluster Kubernetes avec les permissions nécessaires pour :
- Lire les ConfigMaps dans son namespace
- Lister les pods dans tous les namespaces
