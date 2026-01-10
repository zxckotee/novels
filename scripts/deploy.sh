#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Novels Platform Deployment Script ===${NC}"

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found!${NC}"
    echo "Copy .env.example to .env and fill in the values"
    exit 1
fi

# Load environment variables
export $(grep -v '^#' .env | xargs)

# Parse arguments
ACTION=${1:-deploy}

case $ACTION in
    build)
        echo -e "${YELLOW}Building Docker images...${NC}"
        docker-compose -f docker-compose.prod.yml build --no-cache
        echo -e "${GREEN}Build completed!${NC}"
        ;;
        
    deploy)
        echo -e "${YELLOW}Starting deployment...${NC}"
        
        # Pull latest images if needed
        docker-compose -f docker-compose.prod.yml pull db
        
        # Build application images
        echo -e "${YELLOW}Building application images...${NC}"
        docker-compose -f docker-compose.prod.yml build
        
        # Stop existing containers
        echo -e "${YELLOW}Stopping existing containers...${NC}"
        docker-compose -f docker-compose.prod.yml down
        
        # Start new containers
        echo -e "${YELLOW}Starting containers...${NC}"
        docker-compose -f docker-compose.prod.yml up -d
        
        # Wait for services to be ready
        echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
        sleep 10
        
        # Check health
        echo -e "${YELLOW}Checking service health...${NC}"
        docker-compose -f docker-compose.prod.yml ps
        
        echo -e "${GREEN}Deployment completed!${NC}"
        ;;
        
    migrate)
        echo -e "${YELLOW}Running database migrations...${NC}"
        docker-compose -f docker-compose.prod.yml exec api ./server migrate up
        echo -e "${GREEN}Migrations completed!${NC}"
        ;;
        
    logs)
        SERVICE=${2:-}
        if [ -n "$SERVICE" ]; then
            docker-compose -f docker-compose.prod.yml logs -f $SERVICE
        else
            docker-compose -f docker-compose.prod.yml logs -f
        fi
        ;;
        
    status)
        echo -e "${YELLOW}Service status:${NC}"
        docker-compose -f docker-compose.prod.yml ps
        ;;
        
    restart)
        SERVICE=${2:-}
        if [ -n "$SERVICE" ]; then
            echo -e "${YELLOW}Restarting $SERVICE...${NC}"
            docker-compose -f docker-compose.prod.yml restart $SERVICE
        else
            echo -e "${YELLOW}Restarting all services...${NC}"
            docker-compose -f docker-compose.prod.yml restart
        fi
        echo -e "${GREEN}Restart completed!${NC}"
        ;;
        
    stop)
        echo -e "${YELLOW}Stopping all services...${NC}"
        docker-compose -f docker-compose.prod.yml down
        echo -e "${GREEN}Services stopped!${NC}"
        ;;
        
    backup)
        echo -e "${YELLOW}Creating database backup...${NC}"
        BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
        docker-compose -f docker-compose.prod.yml exec -T db pg_dump -U $DB_USER $DB_NAME > ./backups/$BACKUP_FILE
        echo -e "${GREEN}Backup created: ./backups/$BACKUP_FILE${NC}"
        ;;
        
    restore)
        BACKUP_FILE=${2:-}
        if [ -z "$BACKUP_FILE" ]; then
            echo -e "${RED}Error: Please specify backup file${NC}"
            echo "Usage: ./deploy.sh restore <backup_file>"
            exit 1
        fi
        echo -e "${YELLOW}Restoring database from $BACKUP_FILE...${NC}"
        docker-compose -f docker-compose.prod.yml exec -T db psql -U $DB_USER $DB_NAME < $BACKUP_FILE
        echo -e "${GREEN}Restore completed!${NC}"
        ;;
        
    *)
        echo "Usage: $0 {build|deploy|migrate|logs|status|restart|stop|backup|restore}"
        echo ""
        echo "Commands:"
        echo "  build     - Build Docker images"
        echo "  deploy    - Full deployment (build + start)"
        echo "  migrate   - Run database migrations"
        echo "  logs      - View logs (optionally specify service)"
        echo "  status    - Show service status"
        echo "  restart   - Restart services (optionally specify service)"
        echo "  stop      - Stop all services"
        echo "  backup    - Create database backup"
        echo "  restore   - Restore database from backup"
        exit 1
        ;;
esac
