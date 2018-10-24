
/*
*  Operating Systems Assignment 2
* Author: Spencer Cornish
* Date: 10/24/2018
*/

// Includes
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

time_t t;

struct Node {  // Stuct for each node of the linked list
    int id;
    struct Node* prev;
    struct Node* next;
};

// Function Prototypes
void appendNode(struct Node* headNode);
struct Node* newNode(int id, struct Node* prev);
void printList(struct Node* n);
void freeAll(struct Node* n);

int main(void) {
    srand(time(0));
    // Create the head node to get started
    printf("making the root\n");

    struct Node* head = newNode(0, NULL);
    printf("Made the root\n");
    appendNode(head);
    printf("Made the second\n");

    appendNode(head);

    appendNode(head);
    appendNode(head);
    appendNode(head);
    appendNode(head);

    printList(head);

    return 0;
}

// Append a new node to the end of the list.
void appendNode(struct Node* headNode) {
    struct Node* iterNode = NULL;
    iterNode = (struct Node*)malloc(sizeof(struct Node));
    iterNode = headNode;

    while (iterNode->next) {
        iterNode = iterNode->next;
    }
    // Construct a new node
    struct Node* addedNode = newNode(rand() % 50, iterNode);

    // set the former last node's next node to the new final node
    iterNode->next = addedNode;
}

// Create a new node, and set the previousNode and ID
struct Node* newNode(int id, struct Node* prev) {
    struct Node* newNode = NULL;
    newNode = (struct Node*)malloc(sizeof(struct Node));
    newNode->id = id;
    newNode->next = NULL;
    newNode->prev = NULL;
    if (prev != NULL) {
        newNode->prev = prev;
    }

    return newNode;
}
void freeAll(struct Node* n) {
    while (n != NULL) {
        n = n->next;
        free(n->prev);
    }
}

void printList(struct Node* n) {
    while (n != NULL) {
        int prevID, nextID = -1;
        if (n->prev) {
            prevID = n->prev->id;
        }
        if (n->next) {
            nextID = n->next->id;
        }
        printf(" ID: %d PREV: %d NEXT: %d \n", n->id, prevID, nextID);
        n = n->next;
    }
}