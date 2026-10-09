import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useMemo,
  useCallback,
} from 'react';
import {
  DebugEntry,
  DebugFilterOptions,
  DebugLoggerStats,
} from '../types/debug';
import { debugLogger } from '../utils/debugLogger';

export interface DebugContextValue {
  entries: DebugEntry[];
  filteredEntries: DebugEntry[];
  unreadErrorsCount: number;
  isOpen: boolean;
  filterOptions: DebugFilterOptions;
  stats: DebugLoggerStats;
  openDrawer: () => void;
  closeDrawer: () => void;
  toggleDrawer: () => void;
  clearLogs: () => void;
  resetUnreadErrors: () => void;
  setFilterOptions: (options: Partial<DebugFilterOptions>) => void;
}

const DebugContext = createContext<DebugContextValue | undefined>(undefined);

export const DebugProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [entries, setEntries] = useState<DebugEntry[]>(() => debugLogger.getEntries());
  const [unreadErrorsCount, setUnreadErrorsCount] = useState<number>(0);
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const isOpenRef = React.useRef(isOpen);
  useEffect(() => {
    isOpenRef.current = isOpen;
  }, [isOpen]);

  const [filterOptions, setFilterOptionsState] = useState<DebugFilterOptions>({
    level: 'all',
    source: 'all',
    search: '',
  });

  // Iscrizione in tempo reale al singleton debugLogger (una sola volta sul mount)
  useEffect(() => {
    const unsubscribe = debugLogger.subscribe((newEntry, allEntries) => {
      setEntries([...allEntries]);

      // Se arriva un record a livello 'error' e il drawer è chiuso, incrementa gli errori non letti
      if (newEntry.level === 'error' && !isOpenRef.current) {
        setUnreadErrorsCount((prev) => prev + 1);
      }
    });

    return () => {
      unsubscribe();
    };
  }, []);

  // Ascolto globale degli errori JavaScript, unhandled rejections e shortcut da tastiera
  useEffect(() => {
    const handleGlobalError = (event: ErrorEvent) => {
      const sourceLocation = event.filename
        ? `${event.filename}:${event.lineno}:${event.colno}`
        : 'window.onerror';

      debugLogger.logRuntimeError({
        error: event.error || event.message,
        message: event.message,
        stack: event.error?.stack,
        source: sourceLocation,
      });
    };

    const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
      let errMsg = 'Unhandled Promise Rejection';
      let stack: string | undefined;

      if (event.reason instanceof Error) {
        errMsg = event.reason.message;
        stack = event.reason.stack;
      } else if (typeof event.reason === 'string') {
        errMsg = event.reason;
      }

      debugLogger.logRuntimeError({
        error: event.reason,
        message: errMsg,
        stack,
        source: 'unhandledrejection',
      });
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      // Scorciatoia Ctrl + Shift + D (o Cmd + Shift + D per macOS)
      if ((event.ctrlKey || event.metaKey) && event.shiftKey && (event.key === 'D' || event.key === 'd')) {
        event.preventDefault();
        setIsOpen((prev) => {
          const next = !prev;
          if (next) {
            setUnreadErrorsCount(0);
          }
          return next;
        });
      }
    };

    if (typeof window !== 'undefined') {
      window.addEventListener('error', handleGlobalError);
      window.addEventListener('unhandledrejection', handleUnhandledRejection);
      window.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      if (typeof window !== 'undefined') {
        window.removeEventListener('error', handleGlobalError);
        window.removeEventListener('unhandledrejection', handleUnhandledRejection);
        window.removeEventListener('keydown', handleKeyDown);
      }
    };
  }, []);

  const openDrawer = useCallback(() => {
    setIsOpen(true);
    setUnreadErrorsCount(0);
  }, []);

  const closeDrawer = useCallback(() => {
    setIsOpen(false);
  }, []);

  const toggleDrawer = useCallback(() => {
    setIsOpen((prev) => {
      const next = !prev;
      if (next) {
        setUnreadErrorsCount(0);
      }
      return next;
    });
  }, []);

  const clearLogs = useCallback(() => {
    debugLogger.clear();
    setEntries([]);
    setUnreadErrorsCount(0);
  }, []);

  const resetUnreadErrors = useCallback(() => {
    setUnreadErrorsCount(0);
  }, []);

  const setFilterOptions = useCallback((options: Partial<DebugFilterOptions>) => {
    setFilterOptionsState((prev) => ({ ...prev, ...options }));
  }, []);

  // Calcolo memoizzato delle entry filtrate
  const filteredEntries = useMemo(() => {
    return debugLogger.getFilteredEntries(filterOptions);
  }, [entries, filterOptions]);

  // Calcolo memoizzato delle statistiche
  const stats = useMemo(() => {
    return debugLogger.getStats();
  }, [entries]);

  const contextValue: DebugContextValue = useMemo(
    () => ({
      entries,
      filteredEntries,
      unreadErrorsCount,
      isOpen,
      filterOptions,
      stats,
      openDrawer,
      closeDrawer,
      toggleDrawer,
      clearLogs,
      resetUnreadErrors,
      setFilterOptions,
    }),
    [
      entries,
      filteredEntries,
      unreadErrorsCount,
      isOpen,
      filterOptions,
      stats,
      openDrawer,
      closeDrawer,
      toggleDrawer,
      clearLogs,
      resetUnreadErrors,
      setFilterOptions,
    ]
  );

  return <DebugContext.Provider value={contextValue}>{children}</DebugContext.Provider>;
};

export const useDebugConsole = (): DebugContextValue => {
  const context = useContext(DebugContext);
  if (!context) {
    throw new Error("useDebugConsole deve essere utilizzato all'interno di un <DebugProvider>");
  }
  return context;
};
