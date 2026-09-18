# Ручное развертывание на Linux-сервере


## 1. Установить докер

```bash

sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg git

sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

```

Тест:

```bash
docker version
docker compose version
```

Автозапуск при желании:

```bash
sudo systemctl enable --now docker
```

## 2. Сам бэк

```bash
cd /opt
sudo mkdir -p /opt/jobboard
sudo chown "$USER":"$USER" /opt/jobboard
cd /opt/jobboard

git clone https://github.com/Malyshenko-Alexander/jobboard-backend_2026 .
```

## 3. .env

```bash
cp .env.example .env
```

## 4. Сборка

```bash
docker compose up --build -d
```

Статус:

```bash
docker compose ps
docker compose logs -f --tail=50
```

## 5. Обновление

```bash
cd /opt/jobboard
git pull
docker compose up --build -d
docker compose ps
```