'use client';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Copy, CheckCircle2, Send, ArrowRight, ArrowLeft } from 'lucide-react';
import { useState } from 'react';
import { CodeBlock } from '@/components/Layout/code-block';

interface EventField {
  name: string;
  type: string;
  required: boolean;
  description: string;
  example?: any;
}

interface Event {
  id: string;
  name: string;
  description: string;
  category: string;
  method: 'send' | 'receive' | 'both';
  fields: EventField[];
  sampleRequest?: any;
  sampleResponse?: any;
  notes?: string[];
}

interface EventDetailsProps {
  event: Event;
}

export function EventDetails({ event }: EventDetailsProps) {
  const [copiedSection, setCopiedSection] = useState<string | null>(null);

  const copyToClipboard = async (text: string, section: string) => {
    await navigator.clipboard.writeText(text);
    setCopiedSection(section);
    setTimeout(() => setCopiedSection(null), 2000);
  };

  const getMethodIcon = (method: string) => {
    switch (method) {
      case 'send':
        return <Send className="w-4 h-4" />;
      case 'receive':
        return <ArrowLeft className="w-4 h-4" />;
      case 'both':
        return <ArrowRight className="w-4 h-4" />;
      default:
        return <Send className="w-4 h-4" />;
    }
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

  const getMethodDescription = (method: string) => {
    switch (method) {
      case 'send':
        return 'Client sends this event to server';
      case 'receive':
        return 'Client receives this event from server';
      case 'both':
        return 'Bidirectional event (send & receive)';
      default:
        return 'Event direction not specified';
    }
  };

  return (
    <ScrollArea className="h-full">
      <div className="p-8 max-w-5xl mx-auto">
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-4">
            <Badge className={`${getMethodColor(event.method)} flex items-center gap-1`}>
              {getMethodIcon(event.method)}
              {event.method.toUpperCase()}
            </Badge>
            <Badge variant="outline">{event.category}</Badge>
          </div>
          
          <h1 className="text-3xl font-bold mb-2">{event.name}</h1>
          <p className="text-muted-foreground text-lg mb-2">{event.description}</p>
          <p className="text-sm text-muted-foreground">{getMethodDescription(event.method)}</p>
        </div>

        <Tabs defaultValue="schema" className="space-y-6">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="schema">Data Schema</TabsTrigger>
            <TabsTrigger value="request">Sample Request</TabsTrigger>
            <TabsTrigger value="response">Sample Response</TabsTrigger>
            <TabsTrigger value="notes">Notes</TabsTrigger>
          </TabsList>

          <TabsContent value="schema" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Event Data Structure</CardTitle>
                <CardDescription>
                  Field definitions and data types for this event
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {event.fields.map((field, index) => (
                    <div
                      key={index}
                      className="flex items-start justify-between p-4 border rounded-lg hover:bg-accent/50 transition-colors"
                    >
                      <div className="space-y-1 flex-1">
                        <div className="flex items-center gap-2">
                          <code className="font-mono text-sm bg-muted px-2 py-1 rounded">
                            {field.name}
                          </code>
                          <Badge variant={field.required ? "default" : "secondary"}>
                            {field.required ? "Required" : "Optional"}
                          </Badge>
                        </div>
                        <p className="text-sm text-muted-foreground">{field.description}</p>
                        {field.example && (
                          <code className="text-xs bg-muted px-2 py-1 rounded block w-fit">
                            Example: {JSON.stringify(field.example)}
                          </code>
                        )}
                      </div>
                      <Badge variant="outline" className="ml-4">
                        {field.type}
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="request" className="space-y-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Sample Request</CardTitle>
                  <CardDescription>
                    Example data structure when sending this event
                  </CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => copyToClipboard(JSON.stringify(event.sampleRequest, null, 2), 'request')}
                >
                  {copiedSection === 'request' ? (
                    <CheckCircle2 className="w-4 h-4 mr-2" />
                  ) : (
                    <Copy className="w-4 h-4 mr-2" />
                  )}
                  Copy
                </Button>
              </CardHeader>
              <CardContent>
                <CodeBlock
                  language="json"
                  code={JSON.stringify(event.sampleRequest, null, 2)}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="response" className="space-y-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Sample Response</CardTitle>
                  <CardDescription>
                    Example response data for this event
                  </CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => copyToClipboard(JSON.stringify(event.sampleResponse, null, 2), 'response')}
                >
                  {copiedSection === 'response' ? (
                    <CheckCircle2 className="w-4 h-4 mr-2" />
                  ) : (
                    <Copy className="w-4 h-4 mr-2" />
                  )}
                  Copy
                </Button>
              </CardHeader>
              <CardContent>
                <CodeBlock
                  language="json"
                  code={JSON.stringify(event.sampleResponse, null, 2)}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="notes" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Implementation Notes</CardTitle>
                <CardDescription>
                  Additional information and best practices
                </CardDescription>
              </CardHeader>
              <CardContent>
                {event.notes && event.notes.length > 0 ? (
                  <ul className="space-y-3">
                    {event.notes.map((note, index) => (
                      <li key={index} className="flex items-start gap-3">
                        <div className="w-2 h-2 bg-primary rounded-full mt-2 shrink-0" />
                        <p className="text-sm">{note}</p>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-muted-foreground text-sm">No additional notes available for this event.</p>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </ScrollArea>
  );
}