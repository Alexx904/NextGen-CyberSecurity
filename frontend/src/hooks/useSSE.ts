import { useEffect, useRef, useState, useCallback } from 'react';
import { getStoredAccessToken } from '../api/client';
import { debugLogger } from '../utils/debugLogger';

export interface UseSSEOptions {
  url: string | null;
  ticket?: string;
  enabled?: boolean;
  eventNames?: string[];
  onEvent?: (eventType: string, data: any) => void;
  onError?: (error: Event) => void;
  onOpen?: () => void;
}

export interface UseSSEReturn {
  isConnected: boolean;
  isConnecting: boolean;
  error: string | null;
  close: () => void;
}

const DEFAULT_EVENT_NAMES = [
  'message',
  'device_discovered',
  'progress',
  'complete',
  'finding_discovered',
  'phase',
  'status_change',
  'error',
];

export function useSSE({
  url,
  ticket,
  enabled = true,
  eventNames = DEFAULT_EVENT_NAMES,
  onEvent,
  onError,
  onOpen,
}: UseSSEOptions): UseSSEReturn {
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [isConnecting, setIsConnecting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const eventSourceRef = useRef<EventSource | null>(null);
  const onEventRef = useRef(onEvent);
  const onErrorRef = useRef(onError);
  const onOpenRef = useRef(onOpen);

  onEventRef.current = onEvent;
  onErrorRef.current = onError;
  onOpenRef.current = onOpen;

  const close = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
      setIsConnected(false);
      setIsConnecting(false);
    }
  }, []);

  useEffect(() => {
    if (!enabled || !url) {
      close();
      return;
    }

    // Costruzione URL con iniezione stream-ticket prioritario o token JWT di fallback
    let finalUrl = url;
    if (!finalUrl.includes('token=') && !finalUrl.includes('ticket=')) {
      if (ticket) {
        const separator = finalUrl.includes('?') ? '&' : '?';
        finalUrl = `${finalUrl}${separator}ticket=${encodeURIComponent(ticket)}`;
      } else {
        const token = getStoredAccessToken();
        if (token) {
          const separator = finalUrl.includes('?') ? '&' : '?';
          finalUrl = `${finalUrl}${separator}token=${encodeURIComponent(token)}`;
        }
      }
    }

    setIsConnecting(true);
    setError(null);

    let es: EventSource;
    try {
      es = new EventSource(finalUrl);
      eventSourceRef.current = es;
    } catch (err: any) {
      setIsConnecting(false);
      setError(err?.message || 'Impossibile stabilire connessione SSE');
      return;
    }

    es.onopen = () => {
      setIsConnected(true);
      setIsConnecting(false);
      setError(null);
      if (onOpenRef.current) {
        onOpenRef.current();
      }
    };

    es.onerror = (evt: Event) => {
      setIsConnected(false);
      setIsConnecting(false);
      setError('Connessione SSE interrotta o errore di rete');
      if (onErrorRef.current) {
        onErrorRef.current(evt);
      }
    };

    const handleEvent = (eventType: string, evt: MessageEvent) => {
      let parsedData: any = evt.data;
      try {
        parsedData = JSON.parse(evt.data);
      } catch {
        // Se non è JSON valido, mantieni la stringa raw
      }

      // Traccia l'evento su debugLogger
      debugLogger.logSSEEvent({
        eventType,
        data: parsedData,
        url: finalUrl,
        raw: evt.data,
      });

      if (onEventRef.current) {
        onEventRef.current(eventType, parsedData);
      }
    };

    // Registra listener per tutti i tipi di eventi specificati
    const registeredListeners: Array<{ type: string; listener: (e: any) => void }> = [];

    eventNames.forEach((eventType) => {
      const listener = (evt: MessageEvent) => handleEvent(eventType, evt);
      es.addEventListener(eventType, listener);
      registeredListeners.push({ type: eventType, listener });
    });

    // Se tra gli eventi c'è anche il default message
    es.onmessage = (evt: MessageEvent) => handleEvent('message', evt);

    return () => {
      registeredListeners.forEach(({ type, listener }) => {
        es.removeEventListener(type, listener);
      });
      es.close();
      if (eventSourceRef.current === es) {
        eventSourceRef.current = null;
      }
      setIsConnected(false);
      setIsConnecting(false);
    };
  }, [url, enabled, eventNames, close]);

  return {
    isConnected,
    isConnecting,
    error,
    close,
  };
}

