# Arch Stats Overview

This project helps archers improve their performance by collecting and analyzing data about their shots, equipment, and other related conditions. The system uses both hardware and software to gather and process this data, which will be presented for personalized insights and tracking progress over time.

## How It Works

* Registering Archers:
  * Archers sign up using Google OAuth.
  * They provide basic information, such as:
      Full name, email, gender, type of archer (e.g., compound, traditional, barebow), and bow poundage.
  * Archers also register their arrows:
      Information includes length, weight, diameter, spine (flexibility), and a custom name for each arrow.
      Each arrow is programmed with an NFC tag for identification.
* Creating Tournaments:
    Anyone can open a tournament when there isn't one already open.
    Details include the number of lanes and the maximum number of archers per lane.
* Registering for Tournaments:
    Archers can sign up for tournaments and choose an available lane.
    Information about targets (size, scoring zones, and height) is also set up during registration.
* Tracking Shots:
    During a tournament, data about each shot is collected using sensors and saved in real-time.
    Data includes:
        Times when the arrow is engaged, released, and lands.
        The arrow's landing position (x and y coordinates), pull length, and distance to the target.
* Ending Tournaments:
    Once all archers have finished shooting, the tournament is closed, and final results are saved.

## The System Components

* Hardware:
    Raspberry Pi, Bluetooth Low Energy (BLE) beacons, NFC tags, and other sensors to capture shot details.
* Software:
    A Rust app handles real-time data collection and database updates.
    A web server (Python) and Angular app provide user interfaces for archers and administrators.

## Data Stored in the System

### Archers

Personal details like name, email, and bow specifications.
Registered arrows, with each arrow's physical properties and NFC tag information.

### Tournaments

Start and end times.
Number of lanes and participating archers.

### Lanes

Lane numbers and the number of archers per lane.

### Targets

Positioning data (coordinates, height).
Scoring zones (points and radii).
Linked to specific lanes and archers.

### Shots

Recorded data for every arrow shot, such as:
    Timing (engage, release, landing).
    Landing position (coordinates).
    Distance and pull length.

## Planned Features

Initially, the system will handle one lane and one archer for testing. In the future:

Support for multiple lanes and archers.
Integration of networked Raspberry Pis to handle data collection at scale.
Advanced analytics and visualizations to help archers analyze trends in their performance.

## Simple Example

Imagine you're an archer preparing for a tournament:

You register your profile and arrows in the system.
You join a tournament and select a lane.
During the tournament, sensors record every shot you make, including where your arrow lands and how far it traveled.
After the tournament, you can review your performance data, analyze trends, and plan improvements for the next event.
