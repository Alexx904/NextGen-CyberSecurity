import React from 'react';
import { NewScanModal } from './NewScanModal';

interface GroupScanModalProps {
  isOpen: boolean;
  onClose: () => void;
  groupId: string;
  groupName: string;
  targetCount: number;
  onScanStarted: () => void;
}

// Share the inventory-bound authorization flow with the main scan dialog.
export const GroupScanModal: React.FC<GroupScanModalProps> = ({ isOpen, onClose, groupId, onScanStarted }) => (
  <NewScanModal isOpen={isOpen} onClose={onClose} defaultGroupId={groupId} onScanLaunched={onScanStarted} />
);
