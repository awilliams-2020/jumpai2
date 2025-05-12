import { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  TextField,
  Typography,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import client from '../api/client';

interface SchedulingWindow {
  id: number;
  start_hour: number;
  end_hour: number;
  weekday: number;
  is_active: boolean;
}

const weekdays = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
];

export default function SchedulingWindows() {
  const [windows, setWindows] = useState<SchedulingWindow[]>([]);
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({
    start_hour: 9,
    end_hour: 17,
    weekday: 1,
  });
  const [error, setError] = useState<string | null>(null);

  const fetchWindows = async () => {
    try {
      const response = await client.get('/api/scheduling/windows');
      setWindows(response.data);
    } catch (error) {
      console.error('Failed to fetch scheduling windows:', error);
      setError('Failed to load scheduling windows');
    }
  };

  useEffect(() => {
    fetchWindows();
  }, []);

  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
    setFormData({
      start_hour: 9,
      end_hour: 17,
      weekday: 1,
    });
    setError(null);
  };

  const handleSubmit = async () => {
    try {
      const response = await client.post('/api/scheduling/windows', formData);
      setWindows([...windows, response.data]);
      handleClose();
    } catch (error) {
      console.error('Failed to create scheduling window:', error);
      setError('Failed to create scheduling window');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await client.delete(`/api/scheduling/windows/${id}`);
      setWindows(windows.filter(window => window.id !== id));
    } catch (error) {
      console.error('Failed to delete scheduling window:', error);
      setError('Failed to delete scheduling window');
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h4" component="h1">Scheduling Windows</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleClickOpen}
        >
          Add Window
        </Button>
      </Box>

      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(3, 1fr)' },
        gap: 2 
      }}>
        {windows.map((window) => (
          <Card key={String(window.id)}>
            <CardContent>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Box>
                  <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    {weekdays[window.weekday]}
                    {window.is_active && (
                      <Box
                        component="span"
                        sx={{
                          width: 8,
                          height: 8,
                          borderRadius: '50%',
                          bgcolor: 'success.main',
                          display: 'inline-block'
                        }}
                      />
                    )}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {window.start_hour}:00 - {window.end_hour}:00
                  </Typography>
                </Box>
                <IconButton
                  size="small"
                  onClick={() => handleDelete(window.id)}
                  color="error"
                >
                  <DeleteIcon />
                </IconButton>
              </Box>
            </CardContent>
          </Card>
        ))}
      </Box>

      <Dialog open={open} onClose={handleClose}>
        <DialogTitle>Add Scheduling Window</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            <FormControl fullWidth>
              <InputLabel>Weekday</InputLabel>
              <Select
                value={formData.weekday}
                label="Weekday"
                onChange={(e) => setFormData({ ...formData, weekday: Number(e.target.value) })}
              >
                {weekdays.map((day, index) => (
                  <MenuItem key={day} value={index}>
                    {day}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              label="Start Hour"
              type="number"
              value={formData.start_hour}
              onChange={(e) => setFormData({ ...formData, start_hour: Number(e.target.value) })}
              inputProps={{ min: 0, max: 23 }}
              fullWidth
            />

            <TextField
              label="End Hour"
              type="number"
              value={formData.end_hour}
              onChange={(e) => setFormData({ ...formData, end_hour: Number(e.target.value) })}
              inputProps={{ min: 0, max: 23 }}
              fullWidth
            />

            {error && (
              <Typography color="error" variant="body2">
                {error}
              </Typography>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button
            onClick={handleSubmit}
            variant="contained"
            disabled={formData.start_hour >= formData.end_hour}
          >
            Add
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
} 