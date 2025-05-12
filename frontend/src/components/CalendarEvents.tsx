import React, { useEffect, useState } from 'react';
import { format, isValid, parseISO } from 'date-fns';
import {
  Card,
  CardContent,
  Typography,
  CircularProgress,
  Alert,
  Box,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import client from '../api/client';

export interface CalendarEvent {
  id: string;
  summary: string;
  description?: string;
  start_time: string;
  end_time: string;
  location?: string;
  status: string;
  calendar_id: string;
}

// Fetch calendar events for the authenticated user
const fetchCalendarEvents = async (startTime: string, endTime: string): Promise<CalendarEvent[]> => {
  try {
    const response = await client.get('/api/google/calendar/events', {
      params: { start_time: startTime, end_time: endTime }
    });
    // Ensure that we are returning the events array from the response
    return response.data.events || []; 
  } catch (error) {
    console.error('Error fetching calendar events:', error);
    return [];
  }
};

export default function CalendarEvents() {
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadEvents = async () => {
      try {
        setLoading(true);
        setError(null);
        const eventData = await fetchCalendarEvents(
          new Date().toISOString(),
          new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString()
        );
        // Ensure eventData is an array before sorting
        if (Array.isArray(eventData)) {
          setEvents(eventData.sort((a, b) => 
            new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
          ));
        } else {
          // Handle the case where eventData is not an array, e.g., set to empty or log error
          console.error("fetchCalendarEvents did not return an array:", eventData);
          setEvents([]);
        }
      } catch (err) {
        setError('Failed to load calendar events');
        console.error('Error loading calendar events:', err);
      } finally {
        setLoading(false);
      }
    };

    loadEvents();
  }, []);

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error" sx={{ m: 2 }}>
        {error}
      </Alert>
    );
  }

  if (events.length === 0) {
    return (
      <Alert severity="info" sx={{ m: 2 }}>
        No upcoming events found
      </Alert>
    );
  }

  return (
    <Card>
      <CardContent>
        <Typography variant="h5" sx={{ mb: 2, fontWeight: 500 }}>
          Upcoming Events
        </Typography>
        <Grid container spacing={2}>
          {events.map((event) => {
            const startTime = parseISO(event.start_time);
            const endTime = parseISO(event.end_time);
            
            return (
              <Card variant="outlined" key={event.id}>
                <CardContent>
                  <Typography variant="subtitle1" gutterBottom noWrap>
                    {event.summary}
                  </Typography>
                  <Typography variant="body2" color="text.primary">
                    {isValid(startTime) && isValid(endTime) ? (
                      <>
                        {format(startTime, 'MMM d, h:mm a')} -{' '}
                        {format(endTime, 'h:mm a')}
                      </>
                    ) : (
                      'Invalid date'
                    )}
                  </Typography>
                  {event.location && (
                    <Typography variant="body2" color="text.secondary" noWrap>
                      📍 {event.location}
                    </Typography>
                  )}
                  {event.description && (
                    <Typography 
                      variant="body2" 
                      color="text.secondary"
                      sx={{
                        display: '-webkit-box',
                        WebkitLineClamp: 2,
                        WebkitBoxOrient: 'vertical',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                      }}
                    >
                      {event.description.length > 100
                        ? event.description.slice(0, 100) + '...'
                        : event.description}
                    </Typography>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </Grid>
      </CardContent>
    </Card>
  );
}; 