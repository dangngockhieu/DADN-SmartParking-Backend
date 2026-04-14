import { Controller } from '@nestjs/common';
import { VehicleLogService } from './vehicle_log.service';

@Controller('vehicle-log')
export class VehicleLogController {
  constructor(private readonly vehicleLogService: VehicleLogService) {}
}
