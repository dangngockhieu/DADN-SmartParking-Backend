import { Controller, Get, Param, ParseIntPipe } from '@nestjs/common';
import { SlotHistoryService } from './slot_history.service';
import { Roles } from '../../authentication/auth/decorators/roles';
import { Role } from '@prisma/client';
import { ApiBearerAuth, ApiTags } from '@nestjs/swagger';

@ApiTags('Slot History')
@Controller('slot-histories')
export class SlotHistoryController {
  constructor(private readonly slotHistoryService: SlotHistoryService) {}

  // Get slot history by slot ID
  @Get('/:slotId')
  @Roles(Role.ADMIN, Role.MANAGER)
  @ApiBearerAuth('access-token')
  async getSlotHistoryBySlotId(@Param('slotId', ParseIntPipe) slotId: number) {
    const history = await this.slotHistoryService.getSlotHistoryBySlotId(slotId);
    return {
      message: 'Lấy lịch sử thành công',
      data: {
        history
      }
    };
  }
}
