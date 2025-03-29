import winston from "winston";

const logger = winston.createLogger({
  level: 'info',
  format: winston.format.printf(info => {
    return `${(info.message as string).split('\n').join('\n       ')}`.trim();
  }),
  defaultMeta: { service: 'user-service' },
  transports: [
    new winston.transports.File({ filename: 'envim.log' }),
  ],
});

export function log(message: string, level?: string) {
  if (!level) {
    logger.log('info', message)
  } else {
    logger.log(level, message)
  }
}
