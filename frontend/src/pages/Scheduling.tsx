import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import {
	Box,
	Button,
	Container,
	Paper,
	Step,
	StepLabel,
	Stepper,
	Typography,
	TextField,
	FormControl,
	FormLabel,
	Stack,
} from '@mui/material';
import CalendarMonthIcon from '@mui/icons-material/CalendarMonth';
import PersonIcon from '@mui/icons-material/Person';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import Calendar from '../components/Calendar';
import client from '../api/client';
import { format } from 'date-fns';

interface SchedulingLink {
	id: number;
	title: string;
	duration: number;
	max_uses?: number;
	expires_at?: string;
	max_days_in_advance: number;
	custom_questions: string[];
}

interface TimeSlot {
	start: Date;
	end: Date;
}

interface FormData {
	email: string;
	linkedin_url: string;
	answers: { [key: string]: string };
}

const steps = ['Select Date & Time', 'Enter Details', 'Confirmation'];

// Add formatUTCTime function
const formatUTCTime = (date: Date) => {
	return new Intl.DateTimeFormat('en-US', {
		weekday: 'short',
		month: 'short',
		day: 'numeric',
		hour: 'numeric',
		minute: '2-digit',
		hour12: true,
		timeZone: 'UTC'
	}).format(date);
};

export default function Scheduling() {
	const { id } = useParams<{ id: string }>();
	const [activeStep, setActiveStep] = useState(0);
	const [link, setLink] = useState<SchedulingLink | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [selectedSlot, setSelectedSlot] = useState<TimeSlot | null>(null);
	const [formData, setFormData] = useState<FormData>({
		email: '',
		linkedin_url: '',
		answers: {},
	});
	const [submitting, setSubmitting] = useState(false);
	const [success, setSuccess] = useState(false);

	useEffect(() => {
		const fetchLink = async () => {
			try {
				const response = await client.get(`/api/scheduling/links/${id}`);
				setLink(response.data);
				// Initialize answers object with empty strings for each question
				const initialAnswers = response.data.custom_questions.reduce((acc: { [key: string]: string }, question: string) => {
					acc[question] = '';
					return acc;
				}, {});
				setFormData(prev => ({ ...prev, answers: initialAnswers }));
			} catch (err) {
				setError('Failed to load scheduling link');
				console.error('Error fetching scheduling link:', err);
			} finally {
				setLoading(false);
			}
		};

		fetchLink();
	}, [id]);

	const handleSlotSelect = (slot: TimeSlot) => {
		setSelectedSlot(slot);
	};

	const handleInputChange = (field: string, value: string) => {
		setFormData(prev => ({
			...prev,
			[field]: value
		}));
	};

	const handleAnswerChange = (question: string, value: string) => {
		setFormData(prev => ({
			...prev,
			answers: {
				...prev.answers,
				[question]: value
			}
		}));
	};

	const isStep1Valid = () => {
		if (!selectedSlot) return false;
		if (!formData.email) return false;
		if (!formData.linkedin_url) return false;
		// Check if all custom questions are answered
		return Object.values(formData.answers).every(answer => answer.trim() !== '');
	};

	const handleFinish = async () => {
		if (!selectedSlot || !link) return;

		setSubmitting(true);
		setError(null);

		try {
			await client.post(`/api/scheduling/links/${id}/meetings`, {
				client_email: formData.email,
				linkedin_url: formData.linkedin_url,
				start_time: selectedSlot.start,
				end_time: selectedSlot.end,
				answers: formData.answers,
			});

			setSuccess(true);
		} catch (err: any) {
			console.error('Failed to create meeting:', err);
			setError(err.response?.data?.error || 'Failed to create meeting. Please try again.');
		} finally {
			setSubmitting(false);
		}
	};

	if (loading) {
		return (
			<Container maxWidth="md">
				<Box sx={{ my: 4, textAlign: 'center' }}>
					<Typography>Loading...</Typography>
				</Box>
			</Container>
		);
	}

	if (error || !link) {
		return (
			<Container maxWidth="md">
				<Box sx={{ my: 4, textAlign: 'center' }}>
					<Typography color="error">{error || 'Scheduling link not found'}</Typography>
				</Box>
			</Container>
		);
	}

	if (success) {
		return (
			<Container maxWidth="md">
				<Box sx={{ my: 4, textAlign: 'center' }}>
					<CheckCircleIcon sx={{ fontSize: 60, color: 'success.main', mb: 2 }} />
					<Typography variant="h5" gutterBottom>
						Meeting Scheduled Successfully!
					</Typography>
					<Typography color="text.secondary" paragraph>
						You will receive a confirmation email shortly.
					</Typography>
				</Box>
			</Container>
		);
	}

	return (
		<Container maxWidth="md">
			<Box sx={{ my: 4 }}>
				<Typography variant="h4" component="h1" gutterBottom align="center">
					{link.title}
				</Typography>
				<Typography variant="subtitle1" gutterBottom align="center" color="text.secondary">
					Schedule Duration: {link.duration} minutes
				</Typography>

				<Paper sx={{ p: 3, my: 4 }}>
					<Stepper activeStep={activeStep} alternativeLabel>
						{steps.map((label) => (
							<Step key={label}>
								<StepLabel>{label}</StepLabel>
							</Step>
						))}
					</Stepper>

					<Box sx={{ mt: 4, textAlign: 'center' }}>
						{activeStep === 0 && (
							<>
								<CalendarMonthIcon sx={{ fontSize: 60, color: 'primary.main', mb: 2 }} />
								<Typography variant="h6" gutterBottom>
									Select a Date and Time
								</Typography>
								<Typography color="text.secondary" paragraph>
									Choose from available time slots
								</Typography>
								<Calendar 
									onSlotSelect={handleSlotSelect} 
									selectedSlot={selectedSlot || undefined} 
									linkId={id || ''}
									duration={link.duration}
								/>
							</>
						)}

						{activeStep === 1 && (
							<>
								<PersonIcon sx={{ fontSize: 60, color: 'primary.main', mb: 2 }} />
								<Typography variant="h6" gutterBottom>
									Enter Your Details
								</Typography>
								<Typography color="text.secondary" paragraph>
									Please provide your information
								</Typography>
								<Stack spacing={3} sx={{ maxWidth: 600, mx: 'auto', mt: 3 }}>
									<TextField
										label="Email"
										type="email"
										value={formData.email}
										onChange={(e) => handleInputChange('email', e.target.value)}
										required
										fullWidth
									/>
									<TextField
										label="LinkedIn URL"
										value={formData.linkedin_url}
										onChange={(e) => handleInputChange('linkedin_url', e.target.value)}
										required
										fullWidth
										helperText="Please provide your LinkedIn profile URL"
									/>
									{link.custom_questions.map((question, index) => (
										<TextField
											key={index}
											label={question}
											value={formData.answers[question]}
											onChange={(e) => handleAnswerChange(question, e.target.value)}
											required
											fullWidth
											multiline
											rows={3}
										/>
									))}
								</Stack>
							</>
						)}

						{activeStep === 2 && (
							<>
								<CheckCircleIcon sx={{ fontSize: 60, color: 'success.main', mb: 2 }} />
								<Typography variant="h6" gutterBottom>
									Confirmation
								</Typography>
								<Typography color="text.secondary" paragraph>
									Review your booking details
								</Typography>
								<Stack spacing={3} sx={{ maxWidth: 600, mx: 'auto', mt: 3 }}>
									<Paper sx={{ p: 3 }}>
										<Typography variant="subtitle1" gutterBottom>
											Meeting Details
										</Typography>
										<Typography>
											Date: {selectedSlot?.start && formatUTCTime(selectedSlot.start)}
										</Typography>
										<Typography>
											Time: {selectedSlot?.start && formatUTCTime(selectedSlot.start)} - {selectedSlot?.end && formatUTCTime(selectedSlot.end)}
										</Typography>
										<Typography>
											Duration: {link.duration} minutes
										</Typography>
									</Paper>

									<Paper sx={{ p: 3 }}>
										<Typography variant="subtitle1" gutterBottom>
											Your Information
										</Typography>
										<Typography>
											Email: {formData.email}
										</Typography>
										<Typography>
											LinkedIn: {formData.linkedin_url}
										</Typography>
									</Paper>

									<Paper sx={{ p: 3 }}>
										<Typography variant="subtitle1" gutterBottom>
											Your Answers
										</Typography>
										{link.custom_questions.map((question, index) => (
											<Box key={index} sx={{ mb: 2 }}>
												<Typography variant="subtitle2" color="text.secondary">
													{question}
												</Typography>
												<Typography>
													{formData.answers[question]}
												</Typography>
											</Box>
										))}
									</Paper>
								</Stack>
							</>
						)}

						<Box sx={{ mt: 4 }}>
							<Button
								disabled={activeStep === 0 || submitting}
								onClick={() => setActiveStep((prev) => prev - 1)}
								sx={{ mr: 1 }}
							>
								Back
							</Button>
							<Button
								variant="contained"
								onClick={activeStep === steps.length - 1 ? handleFinish : () => setActiveStep((prev) => prev + 1)}
								disabled={
									(activeStep === 0 && !selectedSlot) || 
									(activeStep === 1 && !isStep1Valid()) ||
									submitting
								}
							>
								{activeStep === steps.length - 1 ? 'Finish' : 'Next'}
							</Button>
						</Box>
					</Box>
				</Paper>
			</Box>
		</Container>
	);
} 