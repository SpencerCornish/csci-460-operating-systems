
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

// Color codes, because I want to be able to visually track which thread is printing while I debug
#define ANSI_COLOR_MAGENTA "\x1b[35m"
#define ANSI_COLOR_GREEN "\x1b[32m"
#define ANSI_COLOR_YELLOW "\x1b[33m"
#define ANSI_COLOR_CYAN "\x1b[36m"
#define RESET "\x1B[0m"

// Stuct for each node of the linked list
struct Node {
    int id;
    struct Node *prev;
    struct Node *next;
};

// Mutex for accessing the list
pthread_mutex_t list_mutex = PTHREAD_MUTEX_INITIALIZER;

// Length of the list
int list_length = 0;

// Pointer to the head of the linked list
struct Node *head;

// Function Prototypes
void append_node(int id);
struct Node *new_node(int id, struct Node *prev);
void print_list(struct Node *n);
void *producer_1_entry(void *param);
void *producer_2_entry(void *param);
void *consumer_1_entry(void *param);
void *consumer_2_entry(void *param);

int main(void) {
    pthread_t producer_1;
    pthread_t producer_2;
    pthread_t consumer_1;
    pthread_t consumer_2;

    // Seed the random number generator
    srand(time(0));

    // Create the starter nodes
    printf("making the initial nodes\n");
    head = new_node((rand() % 50) + 1, NULL);
    append_node((rand() % 50) + 1);
    append_node((rand() % 50) + 1);

    if (pthread_create(&producer_1, NULL, producer_1_entry, NULL)) {
        fprintf(__stderrp, "Error creating producer thread 1 \n");
        return 1;
    }
    if (pthread_create(&producer_2, NULL, producer_2_entry, NULL)) {
        fprintf(__stderrp, "Error creating producer thread 2 \n");
        return 1;
    }

    if (pthread_create(&consumer_1, NULL, consumer_1_entry, NULL)) {
        fprintf(__stderrp, "Error creating consumer thread 1 \n");
        return 1;
    }
    if (pthread_create(&consumer_2, NULL, consumer_2_entry, NULL)) {
        fprintf(__stderrp, "Error creating consumer thread 2 \n");
        return 1;
    }

    printf("Threads successfully created\n");

    if (pthread_join(producer_1, NULL)) {
        fprintf(__stderrp, "Error joining producer thread 1\n");
        return 2;
    }

    if (pthread_join(producer_2, NULL)) {
        fprintf(__stderrp, "Error joining producer thread 2\n");
        return 2;
    }

    if (pthread_join(consumer_1, NULL)) {
        fprintf(__stderrp, "Error joining producer thread 1\n");
        return 2;
    }

    if (pthread_join(consumer_2, NULL)) {
        fprintf(__stderrp, "Error joining producer thread 1\n");
        return 2;
    }

    return 0;
}

// An entrypoint for producer 1
void *producer_1_entry(void *param) {
    printf("%sproducer thread number 1 started%s\n", ANSI_COLOR_CYAN, RESET);

    // Add an execution limit for the thread, so the program can be fully exercised each run, and provide sensible output.
    int execution_lim = 0;

    while (execution_lim < 50) {
        if (list_length >= 30) {
            printf("%s[P1] list length reached, waiting...%s\n", ANSI_COLOR_CYAN, RESET);
            while (list_length >= 30) {
                continue;
            }
        } else {
            pthread_mutex_lock(&list_mutex);
            printf("%s[P1] appending new node%s\n", ANSI_COLOR_CYAN, RESET);
            // generate an even random number
            int id;
            do {
                id = (rand() % 50) + 1;
            } while (id % 2 != 1);

            append_node(id);
            print_list(head);
            pthread_mutex_unlock(&list_mutex);
            execution_lim++;
        }
    }

    printf("%s[P1] Max execution reached, exiting...%s\n", ANSI_COLOR_CYAN, RESET);

    return NULL;
}

// An entrypoint for producer 2
void *producer_2_entry(void *param) {
    printf("%sproducer thread number 2 started%s\n", ANSI_COLOR_GREEN, RESET);

    // Add an execution limit for the thread, so the program can be fully exercised each run, and provide sensible output.
    int execution_lim = 0;

    while (execution_lim < 50) {
        if (list_length >= 30) {
            printf("%s[P2] list length reached, waiting...%s\n", ANSI_COLOR_GREEN, RESET);
            while (list_length >= 30) {
                continue;
            }
        } else {
            pthread_mutex_lock(&list_mutex);
            printf("%s[P2] appending new node%s\n", ANSI_COLOR_GREEN, RESET);
            // generate an odd random number
            int id;
            do {
                id = (rand() % 50) + 1;
            } while (id % 2 != 0);
            append_node(id);
            print_list(head);
            pthread_mutex_unlock(&list_mutex);
            execution_lim++;
        }
    }
    printf("%s[P2] Max execution reached, exiting...%s\n", ANSI_COLOR_GREEN, RESET);

    return NULL;
}

// An entrypoint for consumer 1
void *consumer_1_entry(void *param) {
    printf("%sconsumer thread number 1 started%s\n", ANSI_COLOR_MAGENTA, RESET);

    while (1) {
        if (list_length < 1) {
            printf("%s[C1] list empty, waiting for new nodes...%s\n", ANSI_COLOR_MAGENTA, RESET);
            while (list_length < 1) {
                continue;
            }
        } else if (head->id % 2 == 1) {
            pthread_mutex_lock(&list_mutex);
            printf("%s[C1]Popped head with ID %d%s\n", ANSI_COLOR_MAGENTA, head->id, RESET);
            struct Node *new_head = head->next;
            // Get rid of the old head node
            free(head);

            if (new_head != NULL) {
                new_head->prev = (struct Node *)malloc(sizeof(struct Node));
            }
            head = new_head;
            list_length--;
            print_list(head);
            pthread_mutex_unlock(&list_mutex);
        }
    }
    return NULL;
}

// An entrypoint for consumer 1
void *consumer_2_entry(void *param) {
    printf("%sconsumer thread number 2 started%s\n", ANSI_COLOR_YELLOW, RESET);

    while (1) {
        if (list_length < 1) {
            printf("%s[C2] list empty, waiting for new nodes...%s\n", ANSI_COLOR_YELLOW, RESET);
            while (list_length < 1) {
                continue;
            }
        } else if (head->id % 2 == 0) {
            pthread_mutex_lock(&list_mutex);
            printf("%s[C2]Popped head with ID %d%s\n", ANSI_COLOR_YELLOW, head->id, RESET);
            struct Node *new_head = head->next;
            // Get rid of the old head node
            free(head);

            if (new_head != NULL) {
                new_head->prev = (struct Node *)malloc(sizeof(struct Node));
            }
            head = new_head;
            list_length--;
            print_list(head);
            pthread_mutex_unlock(&list_mutex);
        }
    }
    return NULL;
}

// Append a new node to the end of the list with given ID
void append_node(int id) {
    // Find the end of the list
    struct Node *iterNode = NULL;
    iterNode = (struct Node *)malloc(sizeof(struct Node));
    iterNode = head;
    if (iterNode != NULL) {
        while (iterNode->next) {
            iterNode = iterNode->next;
        }
    }

    // Construct a new node, and link it to the end of the list
    struct Node *addedNode = new_node(id, iterNode);
    if (iterNode != NULL) {
        iterNode->next = addedNode;
    } else {
        head = addedNode;
    }
}

// Create a new node, and set its prev node and ID
struct Node *new_node(int id, struct Node *prev) {
    struct Node *new_node = NULL;
    new_node = (struct Node *)malloc(sizeof(struct Node));
    new_node->id = id;
    new_node->next = NULL;
    new_node->prev = prev;

    // Allocate an empty space for the root node's "Prev" node
    // This keeps us from reading garbage data later on accident.
    if (prev == NULL) {
        new_node->prev = (struct Node *)malloc(sizeof(struct Node));
    }

    // Increment the length of the list
    list_length++;

    return new_node;
}

// Prints out the list
void print_list(struct Node *n) {
    printf("CURRENT LIST [");
    while (n != NULL) {
        // Get rid of that last, dangling comma when printing
        if (n->next == NULL) {
            printf("%d", n->id);
        } else {
            printf("%d,", n->id);
        }
        n = n->next;
    }
    printf("]\n");
}