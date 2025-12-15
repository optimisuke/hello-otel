import './otel';
import express, { NextFunction, Request, Response } from 'express';
import { logs as otelLogs, SeverityNumber } from '@opentelemetry/api-logs';
import { PrismaClient, Todo } from '@prisma/client';
import { randomUUID } from 'crypto';
import { z } from 'zod';

const app = express();
const prisma = new PrismaClient();
const PORT = Number(process.env.PORT || 3001);
const SERVICE_NAME = process.env.OTEL_SERVICE_NAME || 'todo-api-node';

const otelLogger = otelLogs.getLogger('todo-api-node');

const log = (...args: unknown[]) => {
  // Simple structured-ish log to stdout; OTEL log exporter will ship stdout.
  console.log('[node-api]', ...args);
  const [event, payload] = args;
  otelLogger.emit({
    body: typeof event === 'string' ? event : 'log',
    severityNumber: SeverityNumber.INFO,
    severityText: 'INFO',
    attributes: {
      'service.name': SERVICE_NAME,
      event: typeof event === 'string' ? event : undefined,
      ...(typeof payload === 'object' ? payload : {}),
    },
  });
};

app.use(express.json());

const todoCreateSchema = z.object({
  title: z.string().min(1).max(200),
  description: z.string().optional(),
  completed: z.boolean().optional(),
});

const todoUpdateSchema = todoCreateSchema.partial();

const uuidSchema = z.string().uuid();

const paginationSchema = z.object({
  skip: z.coerce.number().int().min(0).default(0),
  limit: z.coerce.number().int().min(1).max(500).default(100),
});

const toResponse = (todo: Todo) => ({
  id: todo.id,
  title: todo.title,
  description: todo.description ?? null,
  completed: todo.completed,
  created_at: todo.createdAt,
  updated_at: todo.updatedAt,
});

app.get('/health', (_req, res) => {
  log('health_check', { status: 'ok' });
  res.json({ status: 'healthy', service: SERVICE_NAME });
});

app.get('/api/v1/todos', async (req, res, next) => {
  try {
    const { skip, limit } = paginationSchema.parse(req.query);
    const todos = await prisma.todo.findMany({
      skip,
      take: limit,
      orderBy: { createdAt: 'desc' },
    });
    log('todos_list', { count: todos.length, skip, limit });
    res.json(todos.map(toResponse));
  } catch (err) {
    next(err);
  }
});

app.get('/api/v1/todos/:id', async (req, res, next) => {
  try {
    const todoId = uuidSchema.parse(req.params.id);
    const todo = await prisma.todo.findUnique({ where: { id: todoId } });
    if (!todo) {
      log('todo_not_found', { id: todoId, action: 'get' });
      return res.status(404).json({ detail: `Todo with id ${todoId} not found` });
    }
    log('todo_get', { id: todoId });
    res.json(toResponse(todo));
  } catch (err) {
    next(err);
  }
});

app.post('/api/v1/todos', async (req, res, next) => {
  try {
    const data = todoCreateSchema.parse(req.body);
    const todo = await prisma.todo.create({
      data: {
        id: randomUUID(),
        title: data.title,
        description: data.description,
        completed: data.completed ?? false,
      },
    });
    log('todo_created', { id: todo.id, title: todo.title });
    res.status(201).json(toResponse(todo));
  } catch (err) {
    next(err);
  }
});

app.put('/api/v1/todos/:id', async (req, res, next) => {
  try {
    const todoId = uuidSchema.parse(req.params.id);
    const updateData = todoUpdateSchema.parse(req.body);

    const todo = await prisma.todo.findUnique({ where: { id: todoId } });
    if (!todo) {
      log('todo_not_found', { id: todoId, action: 'update' });
      return res.status(404).json({ detail: `Todo with id ${todoId} not found` });
    }

    const updated = await prisma.todo.update({
      where: { id: todoId },
      data: updateData,
    });

    log('todo_updated', { id: todoId, fields: Object.keys(updateData) });
    res.json(toResponse(updated));
  } catch (err) {
    next(err);
  }
});

app.delete('/api/v1/todos/:id', async (req, res, next) => {
  try {
    const todoId = uuidSchema.parse(req.params.id);
    const todo = await prisma.todo.findUnique({ where: { id: todoId } });
    if (!todo) {
      log('todo_not_found', { id: todoId, action: 'delete' });
      return res.status(404).json({ detail: `Todo with id ${todoId} not found` });
    }

    await prisma.todo.delete({ where: { id: todoId } });
    log('todo_deleted', { id: todoId });
    res.status(204).send();
  } catch (err) {
    next(err);
  }
});

app.use((err: unknown, _req: Request, res: Response, _next: NextFunction) => {
  if (err instanceof z.ZodError) {
    return res.status(400).json({ detail: err.issues });
  }

  return res.status(500).json({ detail: (err as Error).message || 'Internal Server Error' });
});

app.listen(PORT, '0.0.0.0', () => {
  console.log(`Node Todo API listening on port ${PORT}`);
});
