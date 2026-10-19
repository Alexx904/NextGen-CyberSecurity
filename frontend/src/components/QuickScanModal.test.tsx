import { beforeEach, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QuickScanModal } from './QuickScanModal';
import { apiClient } from '../api/client';
import { TargetItem } from '../types/target';

vi.mock('../api/client', () => ({apiClient:{get:vi.fn(),post:vi.fn()}}));
const target = {id:'target',value:'192.168.1.20',type:'IP'} as TargetItem;
beforeEach(()=> {
  vi.resetAllMocks();
  vi.mocked(apiClient.get).mockResolvedValue({data:{success:true,data:{items:[{id:'profile',name:'TCP',nmap_args:'-sS',timeout_seconds:60}]}}});
  vi.mocked(apiClient.post).mockResolvedValue({data:{success:true,data:{id:'scan'}}});
});

it('requires consent for a quick scan and sends the confirmed configuration', async ()=> {
  const close=vi.fn();
  render(<MemoryRouter><QuickScanModal isOpen target={target} onClose={close}/></MemoryRouter>);
  await screen.findByText('TCP');
  const launch=screen.getByRole('button',{name:/Avvia/}) as HTMLButtonElement;
  expect(launch.disabled).toBe(true);
  fireEvent.click(launch); expect(apiClient.post).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole('checkbox',{name:/Confermo di possedere/}));
  fireEvent.click(launch);
  await waitFor(()=>expect(close).toHaveBeenCalledOnce());
  expect(apiClient.post).toHaveBeenCalledWith('/scans',{target_id:'target',profile_id:'profile',authorization:{accepted:true,statement:'scan-authorization-v1',targets:[{id:'target',value:'192.168.1.20'}],nmap_args:'-sS',timeout_seconds:60,cron_expression:''}});
});
