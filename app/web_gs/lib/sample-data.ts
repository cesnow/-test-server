export const events = [
  {
    id: 'user-connect',
    name: 'user.connect',
    description: 'Establishes a new user connection to the WebSocket server',
    category: 'User',
    method: 'send' as const,
    fields: [
      {
        name: 'userId',
        type: 'string',
        required: true,
        description: 'Unique identifier for the user',
        example: 'user_123'
      },
      {
        name: 'token',
        type: 'string',
        required: true,
        description: 'Authentication token for the user session',
        example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
      },
      {
        name: 'metadata',
        type: 'object',
        required: false,
        description: 'Additional connection metadata',
        example: { userAgent: 'Mozilla/5.0...', platform: 'web' }
      }
    ],
    sampleRequest: {
      event: 'user.connect',
      data: {
        userId: 'user_123',
        token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c',
        metadata: {
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
          platform: 'web'
        }
      }
    },
    sampleResponse: {
      event: 'user.connected',
      data: {
        success: true,
        sessionId: 'session_456',
        userId: 'user_123',
        connectedAt: '2024-01-20T10:30:00Z'
      }
    },
    notes: [
      'Authentication token must be valid and not expired',
      'Connection will be automatically closed if authentication fails',
      'Metadata is optional but recommended for analytics and debugging',
      'Each user can have multiple concurrent connections'
    ]
  },
  {
    id: 'user-disconnect',
    name: 'user.disconnect',
    description: 'Gracefully disconnects a user from the WebSocket server',
    category: 'User',
    method: 'both' as const,
    fields: [
      {
        name: 'reason',
        type: 'string',
        required: false,
        description: 'Reason for disconnection',
        example: 'user_logout'
      }
    ],
    sampleRequest: {
      event: 'user.disconnect',
      data: {
        reason: 'user_logout'
      }
    },
    sampleResponse: {
      event: 'user.disconnected',
      data: {
        userId: 'user_123',
        sessionId: 'session_456',
        disconnectedAt: '2024-01-20T10:35:00Z',
        reason: 'user_logout'
      }
    },
    notes: [
      'Can be sent by either client or server',
      'Server will clean up all user-related resources',
      'Any active subscriptions will be automatically cancelled'
    ]
  },
  {
    id: 'chat-message-send',
    name: 'chat.message.send',
    description: 'Sends a chat message to a specific room or user',
    category: 'Chat',
    method: 'send' as const,
    fields: [
      {
        name: 'roomId',
        type: 'string',
        required: true,
        description: 'Target room identifier',
        example: 'room_general'
      },
      {
        name: 'message',
        type: 'string',
        required: true,
        description: 'Message content',
        example: 'Hello everyone!'
      },
      {
        name: 'type',
        type: 'string',
        required: false,
        description: 'Message type (text, image, file)',
        example: 'text'
      },
      {
        name: 'replyTo',
        type: 'string',
        required: false,
        description: 'Message ID being replied to',
        example: 'msg_789'
      }
    ],
    sampleRequest: {
      event: 'chat.message.send',
      data: {
        roomId: 'room_general',
        message: 'Hello everyone! How is everyone doing today?',
        type: 'text',
        replyTo: null
      }
    },
    sampleResponse: {
      event: 'chat.message.sent',
      data: {
        messageId: 'msg_123',
        roomId: 'room_general',
        userId: 'user_123',
        message: 'Hello everyone! How is everyone doing today?',
        type: 'text',
        timestamp: '2024-01-20T10:31:00Z',
        status: 'delivered'
      }
    },
    notes: [
      'Messages are automatically moderated for inappropriate content',
      'Maximum message length is 2000 characters',
      'User must be a member of the room to send messages',
      'Rate limiting applies: 10 messages per minute per user'
    ]
  },
  {
    id: 'chat-message-receive',
    name: 'chat.message.receive',
    description: 'Receives a chat message from another user in the room',
    category: 'Chat',
    method: 'receive' as const,
    fields: [
      {
        name: 'messageId',
        type: 'string',
        required: true,
        description: 'Unique message identifier',
        example: 'msg_456'
      },
      {
        name: 'roomId',
        type: 'string',
        required: true,
        description: 'Room where message was sent',
        example: 'room_general'
      },
      {
        name: 'userId',
        type: 'string',
        required: true,
        description: 'User who sent the message',
        example: 'user_789'
      },
      {
        name: 'username',
        type: 'string',
        required: true,
        description: 'Display name of the sender',
        example: 'JohnDoe'
      },
      {
        name: 'message',
        type: 'string',
        required: true,
        description: 'Message content',
        example: 'Hey there!'
      },
      {
        name: 'timestamp',
        type: 'string',
        required: true,
        description: 'ISO timestamp when message was sent',
        example: '2024-01-20T10:32:00Z'
      }
    ],
    sampleRequest: null,
    sampleResponse: {
      event: 'chat.message.receive',
      data: {
        messageId: 'msg_456',
        roomId: 'room_general',
        userId: 'user_789',
        username: 'JohnDoe',
        message: 'Hey there! Thanks for the warm welcome!',
        type: 'text',
        timestamp: '2024-01-20T10:32:00Z',
        replyTo: 'msg_123'
      }
    },
    notes: [
      'Only sent to users who are subscribed to the room',
      'Messages are delivered in real-time',
      'Includes sender information for display purposes'
    ]
  },
  {
    id: 'game-state-update',
    name: 'game.state.update',
    description: 'Broadcasts current game state to all connected players',
    category: 'Game',
    method: 'receive' as const,
    fields: [
      {
        name: 'gameId',
        type: 'string',
        required: true,
        description: 'Unique game session identifier',
        example: 'game_abc123'
      },
      {
        name: 'state',
        type: 'string',
        required: true,
        description: 'Current game state',
        example: 'playing'
      },
      {
        name: 'players',
        type: 'array',
        required: true,
        description: 'List of active players',
        example: [{ id: 'user_123', name: 'Player1', score: 100 }]
      },
      {
        name: 'round',
        type: 'number',
        required: true,
        description: 'Current round number',
        example: 3
      },
      {
        name: 'timeRemaining',
        type: 'number',
        required: false,
        description: 'Seconds remaining in current round',
        example: 45
      }
    ],
    sampleRequest: null,
    sampleResponse: {
      event: 'game.state.update',
      data: {
        gameId: 'game_abc123',
        state: 'playing',
        players: [
          {
            id: 'user_123',
            name: 'Player1',
            score: 150,
            status: 'active'
          },
          {
            id: 'user_456',
            name: 'Player2',
            score: 120,
            status: 'active'
          }
        ],
        round: 3,
        maxRounds: 5,
        timeRemaining: 45,
        currentTurn: 'user_123'
      }
    },
    notes: [
      'Sent automatically when game state changes',
      'All players in the game receive this update simultaneously',
      'State can be: waiting, playing, paused, finished',
      'timeRemaining is only present during active rounds'
    ]
  },
  {
    id: 'game-player-action',
    name: 'game.player.action',
    description: 'Sends a player action during gameplay',
    category: 'Game',
    method: 'send' as const,
    fields: [
      {
        name: 'gameId',
        type: 'string',
        required: true,
        description: 'Game session identifier',
        example: 'game_abc123'
      },
      {
        name: 'action',
        type: 'string',
        required: true,
        description: 'Type of action being performed',
        example: 'move'
      },
      {
        name: 'data',
        type: 'object',
        required: true,
        description: 'Action-specific data',
        example: { x: 5, y: 3, direction: 'north' }
      }
    ],
    sampleRequest: {
      event: 'game.player.action',
      data: {
        gameId: 'game_abc123',
        action: 'move',
        data: {
          x: 5,
          y: 3,
          direction: 'north',
          timestamp: '2024-01-20T10:33:00Z'
        }
      }
    },
    sampleResponse: {
      event: 'game.action.processed',
      data: {
        gameId: 'game_abc123',
        playerId: 'user_123',
        action: 'move',
        success: true,
        result: {
          newPosition: { x: 5, y: 4 },
          scoreChange: 10
        }
      }
    },
    notes: [
      'Actions are validated server-side before processing',
      'Invalid actions will return an error response',
      'Some actions may trigger immediate state updates',
      'Turn-based games enforce turn order'
    ]
  },
  {
    id: 'notification-push',
    name: 'notification.push',
    description: 'Receives push notifications from the server',
    category: 'System',
    method: 'receive' as const,
    fields: [
      {
        name: 'id',
        type: 'string',
        required: true,
        description: 'Unique notification identifier',
        example: 'notif_xyz789'
      },
      {
        name: 'type',
        type: 'string',
        required: true,
        description: 'Notification type',
        example: 'info'
      },
      {
        name: 'title',
        type: 'string',
        required: true,
        description: 'Notification title',
        example: 'Welcome!'
      },
      {
        name: 'message',
        type: 'string',
        required: true,
        description: 'Notification content',
        example: 'Thanks for joining our platform!'
      },
      {
        name: 'priority',
        type: 'string',
        required: false,
        description: 'Notification priority level',
        example: 'medium'
      }
    ],
    sampleRequest: null,
    sampleResponse: {
      event: 'notification.push',
      data: {
        id: 'notif_xyz789',
        type: 'achievement',
        title: 'Achievement Unlocked!',
        message: 'Congratulations! You have completed your first game.',
        priority: 'high',
        timestamp: '2024-01-20T10:34:00Z',
        actions: [
          {
            label: 'View Achievement',
            action: 'view_achievement',
            data: { achievementId: 'first_game' }
          }
        ]
      }
    },
    notes: [
      'Notifications can include interactive actions',
      'Priority affects display styling and persistence',
      'Types include: info, warning, error, success, achievement',
      'Client should acknowledge receipt of high-priority notifications'
    ]
  }
];