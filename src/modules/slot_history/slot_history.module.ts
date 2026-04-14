import { Module } from '@nestjs/common';
import { SlotHistoryService } from './slot_history.service';
import { SlotHistoryController } from './slot_history.controller';

@Module({
  controllers: [SlotHistoryController],
  providers: [SlotHistoryService],
  exports: [SlotHistoryService]
})
export class SlotHistoryModule {}
