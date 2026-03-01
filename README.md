# DispatchPro - Sistema Logístico

Sistema profesional de gestión logística construido con Go, MongoDB y Bootstrap 5.

## Características

- **Gestión de Pedidos**: CRUD completo con estados
- **Control de Inventario**: Stock con alertas de nivel bajo
- **Gestión de Repartidores**: Control de conductores
- **Asignación de Pedidos**: Asigna pedidos a repartidores
- **Dashboard Analítico**: Métricas en tiempo real
- **Control de Concurrencia**: Mutex para operaciones de stock

## Tecnologías

- Go 1.21, MongoDB 7.0, Bootstrap 5, Chart.js
- Server Side Rendering con html/template
- JWT Authentication
- Control de concurrencia con Mutex

## Requisitos

- Go 1.21+
- MongoDB 7.0 (local o remoto)

## Instalación y Ejecución Local

```bash
# Clonar el repositorio
git clone <repo-url>
cd Programacion-estructurada-c-sharp

# Asegúrate de que MongoDB esté ejecutándose en localhost:27017
# O actualiza MONGODB_URI en el archivo .env

# Instalar dependencias
go mod tidy

# Ejecutar el servidor
go run cmd/server/main.go

# La aplicación estará disponible en http://localhost:8081
```

### Con Docker (Alternativo)

```bash
docker-compose up --build
```

## Colecciones MongoDB

- products - Productos
- orders - Pedidos
- drivers - Repartidores
- inventory_logs - Historial de inventario
- routes - Rutas de entrega
- users - Usuarios del sistema

## API Endpoints

### Autenticación
- `POST /api/auth/register` - Registrar usuario
- `POST /api/auth/login` - Iniciar sesión

### Productos
- `GET /api/products` - Listar productos
- `POST /api/products` - Crear producto
- `GET /api/products/:id` - Obtener producto
- `PUT /api/products/:id` - Actualizar producto
- `POST /api/products/:id/stock` - Ajustar stock
- `GET /api/products/low-stock` - Productos con stock bajo

### Pedidos
- `GET /api/orders` - Listar pedidos
- `POST /api/orders` - Crear pedido
- `GET /api/orders/:id` - Obtener pedido
- `PUT /api/orders/:id/status` - Actualizar estado
- `POST /api/orders/:id/assign` - Asignar repartidor
- `GET /api/orders/stats` - Estadísticas

### Repartidores
- `GET /api/drivers` - Listar repartidores
- `POST /api/drivers` - Crear repartidor
- `GET /api/drivers/:id` - Obtener repartidor
- `PUT /api/drivers/:id` - Actualizar repartidor

## Estructura del Proyecto

```
/
├── cmd/server/           # Punto de entrada
├── internal/
│   ├── config/         # Configuración
│   ├── handlers/       # Controladores HTTP
│   ├── services/      # Lógica de negocio
│   ├── repositories/  # Acceso a datos
│   ├── models/        # Modelos de datos
│   ├── middlewares/   # Middlewares HTTP
│   └── utils/         # Utilidades
├── web/
│   ├── templates/     # Plantillas HTML
│   └── static/        # CSS, JS
├── .env
├── .env.example
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Seguridad

- JWT Authentication
- Roles: admin, user
- Rate limiting
- Hash de contraseñas con bcrypt

---

## 👨‍💻 Desarrollado por Isaac Esteban Haro Torres

**Ingeniero en Sistemas · Full Stack · Automatización · Data**

- 📧 Email: zackharo1@gmail.com
- 📱 WhatsApp: 098805517
- 💻 GitHub: https://github.com/ieharo1
- 🌐 Portafolio: https://ieharo1.github.io/portafolio-isaac.haro/

---

© 2026 Isaac Esteban Haro Torres - Todos los derechos reservados.
