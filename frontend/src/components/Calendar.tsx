import { useState, useEffect } from 'react';
import {
	Box,
	Button,
	Paper,
	Typography,
	CircularProgress,
	Alert,
	Grid,
} from '@mui/material';
import { DateCalendar } from '@mui/x-date-pickers/DateCalendar';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { format, isSameDay } from 'date-fns';
import client from '../api/client';

interface TimeSlot {
	start: Date;
	end: Date;
}

interface CalendarProps {
	onSlotSelect: (slot: TimeSlot) => void;
	selectedSlot?: TimeSlot;
	linkId: string;
	duration: number; // Duration in minutes
}

export default function Calendar({ onSlotSelect, selectedSlot, linkId, duration }: CalendarProps) {
	const [selectedDate, setSelectedDate] = useState<Date | null>(new Date());
	const [availableSlots, setAvailableSlots] = useState<TimeSlot[]>([]);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const fetchAvailableSlots = async (date: Date) => {
		setLoading(true);
		setError(null);
		try {
			const response = await client.get(`/api/scheduling/links/${linkId}/slots`, {
				params: {
					date: format(date, 'yyyy-MM-dd'),
				},
			});
			const slots = response.data.map((slot: any) => ({
				start: new Date(slot.start),
				end: new Date(slot.end),
			}));
			setAvailableSlots(slots);
		} catch (err: any) {
			console.error('Error fetching available slots:', err);
			setError(err.response?.data?.error || 'Failed to load available time slots');
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		if (selectedDate) {
			fetchAvailableSlots(selectedDate);
		}
	}, [selectedDate, linkId]);

	const handleDateSelect = (date: Date | null) => {
		setSelectedDate(date);
	};

	const handleTimeSelect = (slot: TimeSlot) => {
		onSlotSelect(slot);
	};

	const isSlotAvailable = (date: Date) => {
		// Only disable dates in the past
		return date >= new Date(new Date().setHours(0, 0, 0, 0));
	};

	const isSlotSelected = (slot: TimeSlot) => {
		return selectedSlot && 
			selectedSlot.start.getTime() === slot.start.getTime() && 
			selectedSlot.end.getTime() === slot.end.getTime();
	};

	const formatTimeSlot = (date: Date) => {
		// Convert UTC date to local time
		const localDate = new Date(date.getTime() + date.getTimezoneOffset() * 60000);
		return format(localDate, 'h:mm a');
	};

	return (
		<LocalizationProvider dateAdapter={AdapterDateFns}>
			<Box sx={{ flexGrow: 1 }}>
				<Grid container spacing={3}>
					<Grid item xs={12} md={6}>
						<Paper sx={{ p: 2 }}>
							<DateCalendar
								value={selectedDate}
								onChange={handleDateSelect}
								minDate={new Date()}
							/>
						</Paper>
					</Grid>
					<Grid item xs={12} md={6}>
						<Paper sx={{ p: 2 }}>
							<Typography variant="h6" gutterBottom>
								Available Time Slots
							</Typography>
							<Box sx={{ mt: 2 }}>
								{loading ? (
									<Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
										<CircularProgress />
									</Box>
								) : error ? (
									<Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>
								) : selectedDate ? (
									availableSlots.length > 0 ? (
										<Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
											{availableSlots.map((slot, index) => (
												<Button
													key={index}
													variant={isSlotSelected(slot) ? 'contained' : 'outlined'}
													onClick={() => handleTimeSelect(slot)}
													sx={{ minWidth: '120px' }}
												>
													{formatTimeSlot(slot.start)} - {formatTimeSlot(slot.end)}
												</Button>
											))}
										</Box>
									) : (
										<Typography color="text.secondary">
											No available time slots for this date
										</Typography>
									)
								) : (
									<Typography>Please select a date</Typography>
								)}
							</Box>
						</Paper>
					</Grid>
				</Grid>
			</Box>
		</LocalizationProvider>
	);
} 