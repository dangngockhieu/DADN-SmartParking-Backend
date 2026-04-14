import { Controller, Get, Param, ParseIntPipe } from '@nestjs/common';
import { SlotHistoryService } from './slot_history.service';

@Controller('slot-histories')
export class SlotHistoryController {
  constructor(private readonly slotHistoryService: SlotHistoryService) {}

  // Get slot history by slot ID
  @Get('/:slotId')
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
