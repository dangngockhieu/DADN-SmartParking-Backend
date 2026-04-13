declare const module: any;
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import cookieParser from 'cookie-parser';
import { ValidationPipe } from '@nestjs/common';
import { NestExpressApplication } from '@nestjs/platform-express';
import { AppExceptionFilter } from './common/filters/app.exception.filter';
import { TransformInterceptor } from './common/filters/transform.interceptor';
import { join } from 'path';
const bootstrap = async() =>{
  const app = await NestFactory.create<NestExpressApplication>(AppModule);
  app.useStaticAssets(join(__dirname, '..', '..', 'public'));
  app.use(cookieParser());
  const origins = process.env.CORS_ORIGINS?.split(',').map(s => s.trim());
  app.enableCors({
    origin: origins,
    credentials: true
  });
  app.useGlobalPipes(new ValidationPipe({
    whitelist: true
  }));
  app.useGlobalInterceptors(new TransformInterceptor());
  app.useGlobalFilters(new AppExceptionFilter());

  app.setGlobalPrefix('api/v1');

  await app.listen(process.env.PORT ?? 8080);
  if (module.hot) {
    module.hot.accept();
    module.hot.dispose(() => app.close());
  }
}
bootstrap();
