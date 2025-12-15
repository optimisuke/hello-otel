FROM node:20-bullseye-slim

WORKDIR /usr/src/app

RUN apt-get update \
    && apt-get install -y --no-install-recommends libssl1.1 ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY package.json package-lock.json* tsconfig.json prisma ./

RUN npm ci

COPY src ./src

RUN npx prisma generate
RUN npm run build

ENV NODE_ENV=production

EXPOSE 3001
CMD ["node", "-r", "./dist/otel.js", "./dist/index.js"]
