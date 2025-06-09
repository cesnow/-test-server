'use client';

import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { ChevronRight, Folder } from 'lucide-react';
import { useState } from 'react';

interface Event {
  id: string;
  name: string;
  description: string;
  category: string;
  method: 'send' | 'receive' | 'both';
}

interface SidebarProps {
  events: Event[];
  selectedEvent: Event;
  onEventSelectAction: (event: Event) => void;
}

export function Sidebar({ events, selectedEvent, onEventSelectAction }: SidebarProps) {
  const [expandedCategories, setExpandedCategories] = useState<string[]>(['User', 'Chat', 'Game']);

  const categories = events.reduce((acc, event) => {
    if (!acc[event.category]) {
      acc[event.category] = [];
    }
    acc[event.category].push(event);
    return acc;
  }, {} as Record<string, Event[]>);

  const toggleCategory = (category: string) => {
    setExpandedCategories(prev =>
      prev.includes(category)
        ? prev.filter(c => c !== category)
        : [...prev, category]
    );
  };

  const getMethodColor = (method: string) => {
    switch (method) {
      case 'send':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
      case 'receive':
        return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-300';
      case 'both':
        return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-300';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    }
  };

  return (
    <div className="w-80 border-r bg-card">
      <div className="p-4 border-b">
        <h2 className="font-semibold text-lg">Events</h2>
        <p className="text-sm text-muted-foreground">
          {events.length} event{events.length !== 1 ? 's' : ''} available
        </p>
      </div>
      
      <ScrollArea className="h-[calc(100vh-8rem)]">
        <div className="p-4 space-y-2">
          {Object.entries(categories).map(([category, categoryEvents]) => (
            <div key={category} className="space-y-1">
              <button
                onClick={() => toggleCategory(category)}
                className="flex items-center w-full p-2 text-left hover:bg-accent rounded-md transition-colors"
              >
                <ChevronRight
                  className={cn(
                    "w-4 h-4 mr-2 transition-transform",
                    expandedCategories.includes(category) && "rotate-90"
                  )}
                />
                <Folder className="w-4 h-4 mr-2 text-muted-foreground" />
                <span className="font-medium">{category}</span>
                <Badge variant="secondary" className="ml-auto">
                  {categoryEvents.length}
                </Badge>
              </button>
              
              {expandedCategories.includes(category) && (
                <div className="ml-6 space-y-1">
                  {categoryEvents.map((event) => (
                    <button
                      key={event.id}
                      onClick={() => onEventSelectAction(event)}
                      className={cn(
                        "w-full p-3 text-left rounded-lg transition-all duration-200 border",
                        selectedEvent.id === event.id
                          ? "bg-primary text-primary-foreground border-primary shadow-sm"
                          : "hover:bg-accent hover:border-border border-transparent"
                      )}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-medium text-sm">{event.name}</span>
                        <Badge className={getMethodColor(event.method)} variant="secondary">
                          {event.method}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground line-clamp-2">
                        {event.description}
                      </p>
                    </button>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}
